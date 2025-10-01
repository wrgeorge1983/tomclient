#!/usr/bin/env python3
"""
CLI authentication using OAuth with PKCE (no client secret needed)
This can be integrated into your CLI tool
"""

import base64
import hashlib
import json
import os
import secrets
import socket
import threading
import time
import webbrowser
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import parse_qs, urlencode, urlparse
from typing import Optional

import requests


# Configuration - All settings in one place
class Config:
    """Centralized configuration with environment variable support"""
    
    # OAuth/Duo settings
    DUO_CLIENT_ID = os.environ.get('TOM_DUO_CLIENT_ID', '')
    DUO_BASE_URL = os.environ.get('TOM_DUO_BASE_URL', '').rstrip('/')
    
    # OAuth redirect settings
    REDIRECT_PORT = int(os.environ.get('TOM_OAUTH_REDIRECT_PORT', '8899'))
    REDIRECT_PATH = os.environ.get('TOM_OAUTH_REDIRECT_PATH', '/callback')
    
    # OAuth scopes
    OAUTH_SCOPES = os.environ.get('TOM_OAUTH_SCOPES', 'openid email profile')
    
    # Timeouts
    AUTH_TIMEOUT = int(os.environ.get('TOM_AUTH_TIMEOUT', '120'))  # seconds
    
    # Token storage
    TOKEN_DIR = os.environ.get('TOM_CONFIG_DIR', '~/.tom')
    TOKEN_FILE = os.environ.get('TOM_TOKEN_FILE', 'token.json')
    CONFIG_FILE = os.environ.get('TOM_CONFIG_FILE', 'config.json')
    
    # File permissions (octal)
    FILE_PERMISSIONS = 0o600
    
    @classmethod
    def get_token_path(cls):
        """Get full path to token file"""
        return os.path.join(os.path.expanduser(cls.TOKEN_DIR), cls.TOKEN_FILE)
    
    @classmethod
    def get_config_path(cls):
        """Get full path to config file"""
        return os.path.join(os.path.expanduser(cls.TOKEN_DIR), cls.CONFIG_FILE)


class AuthServer(HTTPServer):
    """HTTPServer with OAuth state storage"""
    auth_code: Optional[str] = None
    state: Optional[str] = None


class OAuthCallbackHandler(BaseHTTPRequestHandler):
    """Handle OAuth callback on localhost"""
    
    def do_GET(self):
        """Handle GET request from OAuth callback"""
        query = parse_qs(urlparse(self.path).query)
        
        if 'code' in query:
            if isinstance(self.server, AuthServer):
                self.server.auth_code = query['code'][0]
                self.server.state = query.get('state', [None])[0]
            
            # Send success response to browser
            self.send_response(200)
            self.send_header('Content-type', 'text/html')
            self.end_headers()
            self.wfile.write(b"""
                <html>
                <body>
                <h1>Authentication Successful!</h1>
                <p>You can close this window and return to your terminal.</p>
                <script>window.setTimeout(function(){window.close()}, 2000);</script>
                </body>
                </html>
            """)
        else:
            # Handle error
            error = query.get('error', ['Unknown error'])[0]
            self.send_response(400)
            self.send_header('Content-type', 'text/html')
            self.end_headers()
            self.wfile.write(f"""
                <html>
                <body>
                <h1>Authentication Failed</h1>
                <p>Error: {error}</p>
                </body>
                </html>
            """.encode())
    
    def log_message(self, format, *args):
        """Suppress log messages"""
        pass


class DuoPKCEAuth:
    """Handle Duo authentication using PKCE flow"""
    
    def __init__(self, client_id=None, duo_base_url=None, redirect_port=None):
        # Use provided values or fall back to config/env vars
        self.client_id = client_id or Config.DUO_CLIENT_ID
        self.duo_base_url = (duo_base_url or Config.DUO_BASE_URL).rstrip('/')
        self.redirect_port = redirect_port or Config.REDIRECT_PORT
        self.redirect_uri = f"http://localhost:{self.redirect_port}{Config.REDIRECT_PATH}"
        
        # Validate required settings
        if not self.client_id:
            raise ValueError("DUO_CLIENT_ID not provided. Set TOM_DUO_CLIENT_ID environment variable or pass client_id parameter")
        if not self.duo_base_url:
            raise ValueError("DUO_BASE_URL not provided. Set TOM_DUO_BASE_URL environment variable or pass duo_base_url parameter")
        
        # Provider-specific endpoint configuration
        if 'duosecurity.com' in self.duo_base_url:
            # Duo: Convert api-XXX URL to sso-XXX URL
            if 'api-' in self.duo_base_url:
                tenant_id = self.duo_base_url.split('api-')[1].split('.')[0]
                oidc_base = f"https://sso-{tenant_id}.sso.duosecurity.com/oidc/{self.client_id}"
            else:
                oidc_base = self.duo_base_url
        elif 'login.microsoftonline.com' in self.duo_base_url:
            # Microsoft: Standard OAuth endpoints
            tenant = self.duo_base_url.split('/')[-1] if '/' in self.duo_base_url else 'common'
            oidc_base = f"https://login.microsoftonline.com/{tenant}/oauth2/v2.0"
        elif 'accounts.google.com' in self.duo_base_url:
            # Google: Standard OAuth endpoints
            oidc_base = "https://accounts.google.com"
        elif 'github.com' in self.duo_base_url:
            # GitHub: Standard OAuth endpoints
            oidc_base = "https://github.com/login/oauth"
        else:
            # Generic OIDC provider
            oidc_base = self.duo_base_url
        
        # OAuth/OIDC endpoints
        self.authorize_url = f"{oidc_base}/authorize"
        self.token_url = f"{oidc_base}/token"
        self.jwks_url = f"{oidc_base}/jwks"
        
        print(f"Using OIDC endpoints:")
        print(f"  Authorize: {self.authorize_url}")
        print(f"  Token: {self.token_url}")
    
    def generate_pkce_params(self):
        """Generate PKCE code verifier and challenge"""
        # Generate code verifier (43-128 characters)
        code_verifier = base64.urlsafe_b64encode(
            secrets.token_bytes(32)
        ).decode('utf-8').rstrip('=')
        
        # Generate code challenge (SHA256 of verifier)
        challenge_bytes = hashlib.sha256(code_verifier.encode()).digest()
        code_challenge = base64.urlsafe_b64encode(
            challenge_bytes
        ).decode('utf-8').rstrip('=')
        
        return code_verifier, code_challenge
    
    def get_free_port(self):
        """Find a free port if the default is taken"""
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            try:
                s.bind(('', self.redirect_port))
                return self.redirect_port
            except:
                # Port is taken, find another
                s.bind(('', 0))
                return s.getsockname()[1]
    
    def start_callback_server(self):
        """Start local HTTP server to receive OAuth callback"""
        port = self.get_free_port()
        if port != self.redirect_port:
            self.redirect_port = port
            self.redirect_uri = f"http://localhost:{port}/callback"
            print(f"Note: Using port {port} for callback")
        
        server = AuthServer(('localhost', port), OAuthCallbackHandler)
        server.auth_code = None
        server.state = None
        server.timeout = Config.AUTH_TIMEOUT
        
        # Run server in background thread
        thread = threading.Thread(target=server.handle_request)
        thread.daemon = True
        thread.start()
        
        return server
    
    def authenticate(self):
        """Perform the full PKCE authentication flow"""
        print("Starting Duo authentication...")
        
        # Generate PKCE parameters
        code_verifier, code_challenge = self.generate_pkce_params()
        state = secrets.token_urlsafe(16)
        
        # Build authorization URL
        auth_params = {
            'client_id': self.client_id,
            'redirect_uri': self.redirect_uri,
            'response_type': 'code',
            'scope': Config.OAUTH_SCOPES,
            'state': state,
            'code_challenge': code_challenge,
            'code_challenge_method': 'S256'
        }
        
        auth_url = f"{self.authorize_url}?{urlencode(auth_params)}"
        
        # Debug: Show what we're sending
        print(f"\nDebug - Authorization URL: {self.authorize_url}")
        print(f"Debug - Parameters being sent:")
        for key, value in auth_params.items():
            if key == 'code_challenge':
                print(f"  {key}: {value[:20]}...")
            else:
                print(f"  {key}: {value}")
        
        # Start callback server
        print(f"\nStarting callback server on {self.redirect_uri}")
        server = self.start_callback_server()
        
        # Open browser for user to authenticate
        print(f"\nOpening browser for authentication...")
        print(f"If the browser doesn't open, visit:\n{auth_url}\n")
        webbrowser.open(auth_url)
        
        # Wait for callback
        print(f"Waiting for authentication (timeout: {Config.AUTH_TIMEOUT} seconds)...")
        start_time = time.time()
        while server.auth_code is None and (time.time() - start_time) < Config.AUTH_TIMEOUT:
            time.sleep(0.5)
        
        if server.auth_code is None:
            raise Exception("Authentication timed out")
        
        # Verify state
        if server.state != state:
            raise Exception("State mismatch - possible CSRF attack")
        
        print("Authorization code received, exchanging for token...")
        
        # Exchange code for token using PKCE (no client_secret!)
        token_data = {
            'grant_type': 'authorization_code',
            'code': server.auth_code,
            'client_id': self.client_id,
            'redirect_uri': self.redirect_uri,
            'code_verifier': code_verifier  # This replaces client_secret!
        }
        
        response = requests.post(self.token_url, data=token_data)
        
        if response.status_code != 200:
            raise Exception(f"Token exchange failed: {response.text}")
        
        token_response = response.json()
        return token_response
    
    def save_token(self, token_data, config_file=None):
        """Save token to local config file"""
        config_path = config_file or Config.get_token_path()
        config_dir = os.path.dirname(config_path)
        
        # Ensure directory exists
        os.makedirs(config_dir, exist_ok=True)
        
        # Add metadata
        token_data['obtained_at'] = time.time()
        token_data['expires_at'] = time.time() + token_data.get('expires_in', 3600)
        
        with open(config_path, 'w') as f:
            json.dump(token_data, f, indent=2)
        
        # Secure the file
        os.chmod(config_path, Config.FILE_PERMISSIONS)
        
        print(f"Token saved to {config_path}")
    
    def load_token(self, config_file=None):
        """Load saved token if valid"""
        config_path = config_file or Config.get_token_path()
        
        if not os.path.exists(config_path):
            return None
        
        try:
            with open(config_path, 'r') as f:
                token_data = json.load(f)
            
            # Check if token is expired
            if time.time() >= token_data.get('expires_at', 0):
                print("Saved token is expired")
                return None
            
            return token_data
        except Exception as e:
            print(f"Error loading token: {e}")
            return None


def main():
    """Example usage"""
    # Configuration can come from:
    # 1. Environment variables (TOM_DUO_CLIENT_ID, TOM_DUO_BASE_URL)
    # 2. Command line arguments
    # 3. Config file
    
    # Try to create auth client (will use env vars if set)
    try:
        auth = DuoPKCEAuth()
    except ValueError as e:
        print(f"Configuration error: {e}")
        print("\nPlease set the following environment variables:")
        print("  export TOM_DUO_CLIENT_ID='your-client-id'")
        print("  export TOM_DUO_BASE_URL='https://your-tenant.duosecurity.com'")
        return 1
    
    # Try to load existing token
    token_data = auth.load_token()
    
    if token_data:
        print("Using saved token")
    else:
        # Authenticate and save token
        try:
            token_data = auth.authenticate()
            auth.save_token(token_data)
            print("\n✅ Authentication successful!")
        except Exception as e:
            print(f"\n❌ Authentication failed: {e}")
            return 1
    
    # Now use the token with your API
    access_token = token_data.get('access_token') or token_data.get('id_token')
    print(f"\nToken type: {token_data.get('token_type', 'Bearer')}")
    print(f"Access token: {access_token[:20]}...")
    
    # Make API call
    # response = requests.get(
    #     "http://localhost:8020/api/",
    #     headers={"Authorization": f"Bearer {access_token}"}
    # )
    
    return 0


if __name__ == "__main__":
    import sys
    sys.exit(main())