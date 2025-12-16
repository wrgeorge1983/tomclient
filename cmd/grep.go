package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"tomclient/auth"
)

var (
	grepDevices     string
	grepFilter      string
	grepMatch       string
	grepCommand     string
	grepContext     int
	grepBefore      int
	grepAfter       int
	grepSection     bool
	grepParent      bool
	grepParentLine  bool
	grepIgnoreCase  bool
	grepNoColor     bool
	grepLineNumbers bool
	grepNoCache     bool
	grepParallel    int
)

type grepResult struct {
	Device  string
	Matches []matchBlock
	Error   error
}

type matchBlock struct {
	Lines      []string
	MatchIndex int // index of the matching line within Lines
}

var grepCmd = &cobra.Command{
	Use:   "grep <pattern>",
	Short: "Search device configs for matching lines",
	Long: `Search device configurations for lines matching a pattern.
Similar to ripgrep, supports context lines and Cisco-style section matching.

Context Modes:
  -C N           Show N lines before and after each match
  -A N           Show N lines after each match  
  -B N           Show N lines before each match
  --section      Show matching line and all indented children (Cisco section-style)
  --parent       Show parent block header and all siblings at same indentation
  --parent-line  Show just the parent header line and the match`,
	Example: `  tomclient grep "ip route.*10.0.0" --devices SCCSNJ75AS1,SCCSNJ75AS2
  tomclient grep "interface.*Loopback" --match "^SCCSNJ" -C 3
  tomclient grep "bgp neighbor" --section --match ".*AS[12]"
  tomclient grep "shutdown" -P --devices SCCSNJ75AS1
  tomclient grep "ip address" -p --match ".*AS1$"
  tomclient grep "permit" --command "show access-lists" -A 2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		pattern := args[0]
		if grepIgnoreCase {
			pattern = "(?i)" + pattern
		}

		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}

		devices, err := resolveDevices()
		if err != nil {
			return err
		}

		if len(devices) == 0 {
			return fmt.Errorf("no devices specified; use --devices, --prefix, --match, or --filter")
		}

		results := queryDevicesParallel(devices, re)
		printResults(results, re)

		return nil
	},
}

func resolveDevices() ([]string, error) {
	if grepDevices != "" {
		return strings.Split(grepDevices, ","), nil
	}

	// If --match is specified, query the API directly with Caption filter
	if grepMatch != "" {
		inventory, err := client.ExportInventoryWithFilters(map[string]string{
			"Caption": grepMatch,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to query inventory: %w", err)
		}

		devices := make([]string, 0, len(inventory))
		for name := range inventory {
			devices = append(devices, name)
		}
		return devices, nil
	}

	cfg, err := auth.LoadConfig(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// If --filter is specified, query the API with named filter
	if grepFilter != "" {
		inventory, err := client.ExportInventory(grepFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to query inventory with filter %q: %w", grepFilter, err)
		}

		devices := make([]string, 0, len(inventory))
		for name := range inventory {
			devices = append(devices, name)
		}
		return devices, nil
	}

	cache, err := auth.LoadInventoryCache(cfg.ConfigDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load inventory cache: %w", err)
	}

	if cache == nil {
		return nil, fmt.Errorf("no inventory cache; run 'tomclient inventory --refresh' first")
	}

	return cache.Devices, nil
}

func queryDevicesParallel(devices []string, re *regexp.Regexp) []grepResult {
	results := make([]grepResult, len(devices))
	var wg sync.WaitGroup

	parallel := grepParallel
	if parallel <= 0 {
		parallel = 10
	}
	sem := make(chan struct{}, parallel)

	for i, device := range devices {
		wg.Add(1)
		go func(idx int, dev string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			results[idx] = queryDevice(dev, re)
		}(i, device)
	}

	wg.Wait()
	return results
}

func queryDevice(device string, re *regexp.Regexp) grepResult {
	command := grepCommand
	if command == "" {
		command = "show running-config"
	}

	output, err := client.SendDeviceCommand(device, command, true, true, !grepNoCache, nil, false)
	if err != nil {
		return grepResult{Device: device, Error: err}
	}

	lines := strings.Split(output, "\n")
	matches := findMatches(lines, re)

	return grepResult{Device: device, Matches: matches}
}

func findMatches(lines []string, re *regexp.Regexp) []matchBlock {
	var matches []matchBlock

	for i, line := range lines {
		if re.MatchString(line) {
			var block matchBlock

			if grepSection {
				block = extractSection(lines, i)
			} else if grepParent {
				block = extractParentBlock(lines, i)
			} else if grepParentLine {
				block = extractParentLine(lines, i)
			} else {
				block = extractContext(lines, i)
			}

			matches = append(matches, block)
		}
	}

	return mergeOverlappingBlocks(matches)
}

func extractContext(lines []string, matchIdx int) matchBlock {
	before := grepBefore
	after := grepAfter
	if grepContext > 0 {
		before = grepContext
		after = grepContext
	}

	start := matchIdx - before
	if start < 0 {
		start = 0
	}

	end := matchIdx + after + 1
	if end > len(lines) {
		end = len(lines)
	}

	return matchBlock{
		Lines:      lines[start:end],
		MatchIndex: matchIdx - start,
	}
}

func extractSection(lines []string, matchIdx int) matchBlock {
	matchLine := lines[matchIdx]
	matchIndent := getIndent(matchLine)

	start := matchIdx
	end := matchIdx + 1

	for end < len(lines) {
		lineIndent := getIndent(lines[end])
		// Empty lines are included
		if strings.TrimSpace(lines[end]) == "" {
			end++
			continue
		}
		// Stop when indentation returns to match level or less
		if lineIndent <= matchIndent {
			break
		}
		end++
	}

	return matchBlock{
		Lines:      lines[start:end],
		MatchIndex: 0,
	}
}

func extractParentBlock(lines []string, matchIdx int) matchBlock {
	matchLine := lines[matchIdx]
	matchIndent := getIndent(matchLine)

	// Find the parent (last line with less indentation before match)
	parentIdx := matchIdx
	for i := matchIdx - 1; i >= 0; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineIndent := getIndent(line)
		if lineIndent < matchIndent {
			parentIdx = i
			break
		}
	}

	// If we're already at root level (no indent), just show context
	if parentIdx == matchIdx && matchIndent == 0 {
		return extractContext(lines, matchIdx)
	}

	parentIndent := getIndent(lines[parentIdx])

	// Find the end of the parent's block
	end := parentIdx + 1
	for end < len(lines) {
		line := lines[end]
		if strings.TrimSpace(line) == "" {
			end++
			continue
		}
		lineIndent := getIndent(line)
		if lineIndent <= parentIndent {
			break
		}
		end++
	}

	return matchBlock{
		Lines:      lines[parentIdx:end],
		MatchIndex: matchIdx - parentIdx,
	}
}

func extractParentLine(lines []string, matchIdx int) matchBlock {
	matchLine := lines[matchIdx]
	matchIndent := getIndent(matchLine)

	// Find the parent (last line with less indentation before match)
	parentIdx := -1
	for i := matchIdx - 1; i >= 0; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineIndent := getIndent(line)
		if lineIndent < matchIndent {
			parentIdx = i
			break
		}
	}

	// If no parent found (already at root level), just return the match
	if parentIdx == -1 {
		return matchBlock{
			Lines:      []string{lines[matchIdx]},
			MatchIndex: 0,
		}
	}

	return matchBlock{
		Lines:      []string{lines[parentIdx], lines[matchIdx]},
		MatchIndex: 1,
	}
}

func getIndent(line string) int {
	indent := 0
	for _, ch := range line {
		if ch == ' ' {
			indent++
		} else if ch == '\t' {
			indent += 4
		} else {
			break
		}
	}
	return indent
}

func mergeOverlappingBlocks(blocks []matchBlock) []matchBlock {
	if len(blocks) <= 1 {
		return blocks
	}

	// For now, just return as-is; could implement merging later
	// to avoid duplicate output when matches are close together
	return blocks
}

func printResults(results []grepResult, re *regexp.Regexp) {
	deviceColor := color.New(color.FgMagenta, color.Bold)
	matchColor := color.New(color.FgRed, color.Bold)
	lineNumColor := color.New(color.FgGreen)
	sepColor := color.New(color.FgCyan)

	if grepNoColor {
		color.NoColor = true
	}

	first := true
	for _, result := range results {
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "%s: error: %v\n", result.Device, result.Error)
			continue
		}

		if len(result.Matches) == 0 {
			continue
		}

		if !first {
			fmt.Println()
		}
		first = false

		deviceColor.Printf("=== %s ===\n", result.Device)

		for blockIdx, block := range result.Matches {
			if blockIdx > 0 {
				sepColor.Println("--")
			}

			for i, line := range block.Lines {
				if grepLineNumbers {
					lineNum := fmt.Sprintf("%4d", i+1)
					lineNumColor.Print(lineNum)

					sep := ":"
					if i == block.MatchIndex {
						sep = ">"
					}
					sepColor.Printf("%s ", sep)
				}

				// Highlight matches in the line
				if re.MatchString(line) {
					highlighted := re.ReplaceAllStringFunc(line, func(match string) string {
						return matchColor.Sprint(match)
					})
					fmt.Println(highlighted)
				} else {
					fmt.Println(line)
				}
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(grepCmd)

	// Device selection
	grepCmd.Flags().StringVarP(&grepDevices, "devices", "D", "", "Comma-separated list of device names")
	grepCmd.Flags().StringVarP(&grepMatch, "match", "m", "", "Regex pattern to match device names")
	grepCmd.Flags().StringVarP(&grepFilter, "filter", "f", "", "Use named inventory filter")

	// Command to run
	grepCmd.Flags().StringVar(&grepCommand, "command", "show running-config", "Command to execute on devices")

	// Context modes
	grepCmd.Flags().IntVarP(&grepContext, "context", "C", 0, "Lines of context around matches")
	grepCmd.Flags().IntVarP(&grepBefore, "before", "B", 0, "Lines before each match")
	grepCmd.Flags().IntVarP(&grepAfter, "after", "A", 0, "Lines after each match")
	grepCmd.Flags().BoolVar(&grepSection, "section", false, "Show match and all indented children")
	grepCmd.Flags().BoolVarP(&grepParent, "parent", "P", false, "Show parent block and siblings")
	grepCmd.Flags().BoolVarP(&grepParentLine, "parent-line", "p", false, "Show just the parent line and match")

	// Other options
	grepCmd.Flags().BoolVarP(&grepIgnoreCase, "ignore-case", "i", false, "Case insensitive matching")
	grepCmd.Flags().BoolVar(&grepNoColor, "no-color", false, "Disable colored output")
	grepCmd.Flags().BoolVarP(&grepLineNumbers, "line-numbers", "n", false, "Show line numbers")
	grepCmd.Flags().BoolVar(&grepNoCache, "no-cache", false, "Disable caching (cache enabled by default)")
	grepCmd.Flags().IntVar(&grepParallel, "parallel", 10, "Number of parallel device queries")
}
