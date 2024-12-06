package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
)

var start int
var pline int

var symbols []rune

func main() {
	symbols = []rune("󰞔󰪸󰸞")
	/*
		//  ╭───╮
		//  │  │
		//  ╰───╯
		disk := "nvme0n1"
		parts := getPartitions()
		mounted_parts(disk, parts)
		part_labels(disk, parts)
		fmt.Println(parts)
		for _, val := range parts {
			x, y := get_used_p(disk+"p"+val[0], val[4])
			if y != nil {
				fmt.Println(disk+"p"+val[0], val[4], "--")
			} else {
				fmt.Println(disk+"p"+val[0], val[4], fmt.Sprintf("%.2f%%", x))
			}
		}
		return
		//*/
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Error creating screen: %v", err)
	}
	defer screen.Fini()

	err = screen.Init()
	if err != nil {
		log.Fatalf("Error initializing screen: %v", err)
	}

	screen.Clear()
	refresh(screen)
	screen.Show()

	for {
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Rune() == 'q' {
				return
			}
			if ev.Key() == tcell.KeyCtrlC {
				return
			}
		case *tcell.EventResize:
			screen.Sync()
			screen.Clear()
			refresh(screen)
			screen.Show()
		}
	}
}

func refresh(screen tcell.Screen) {
	start = 0
	pline = 25
	// style := tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FF8040"))
	width, height := screen.Size()
	// screen.SetContent(10, 5, 'H', nil, style)
	// screen.SetContent(11, 5, 'i', nil, style)
	// draw_box(screen, 0, 0, width-1, height-1, style, "═║╔╚╗╝")
	// draw_box(screen, 45, 10, 85, 20, style, "═║╔╚╗╝")
	disk := "nvme0n1"
	parts := getPartitions()
	mounted_parts(disk, parts)
	part_labels(disk, parts)
	empty_parts := 0
	if width > 100 {
		print_at(screen, 3, 1, string(symbols[0]), tcell.StyleDefault.Foreground(tcell.GetColor("#00FF00")))
		print_at(screen, 6, 1, string(symbols[1]), tcell.StyleDefault.Foreground(tcell.GetColor("#FF0000")))
		print_at(screen, 9, 1, "│", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
		print_at(screen, 11, 1, string(symbols[2]), tcell.StyleDefault.Foreground(tcell.GetColor("#FFFFFF")))
		print_at(screen, 14, 1, "│", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
		print_at(screen, 16, 1, string(symbols[3]), tcell.StyleDefault.Foreground(tcell.GetColor("#FFFFAA")))
		print_at(screen, 19, 1, string(symbols[4]), tcell.StyleDefault.Foreground(tcell.GetColor("#FFFFAA")))
		print_at(screen, 22, 1, "│", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
		print_at(screen, 24, 1, string(symbols[5]), tcell.StyleDefault.Foreground(tcell.GetColor("#FFFF00")))
		print_at(screen, 27, 1, string(symbols[6]), tcell.StyleDefault.Foreground(tcell.GetColor("#00FF00")))
		print_head(screen, width, height)
	}
	for _, val := range parts {
		size, err := strconv.ParseFloat(strings.TrimRight(val[3], "GB"), 64)
		if err != nil {
			fmt.Println("Error parsing used space:", err)
			return
		}
		if int(math.Round((size/512)*100)) == 0 {
			empty_parts++
		}
	}
	print_it(screen, fmt.Sprintf("The width is: %d & height is: %d", width, height))
	for i, val := range parts {
		x, ey := get_used_p(disk+"p"+val[0], val[4])
		if ey != nil {
			// fmt.Println(disk+"p"+val[0], val[4], "--")
			size, err := strconv.ParseFloat(strings.TrimRight(val[3], "GB"), 64)
			if err != nil {
				fmt.Println("Error parsing used space:", err)
				return
			}
			// print_it(screen, fmt.Sprint(((size / 512) * 100)))
			draw_it(screen, ((size / 512) * 100), 0, width, val[4], "/dev/"+disk+"p"+val[0], val[3])
		} else {
			// fmt.Println(disk+"p"+val[0], val[4], fmt.Sprintf("%.2f%%", x))
			// print_it(screen, fmt.Sprintf("%.2f%%", x)+":hello:"+fmt.Sprintf("%d", int(math.Round(x))), style)
			size, err := strconv.ParseFloat(strings.TrimRight(val[3], "GB"), 64)
			if err != nil {
				fmt.Println("Error parsing used space:", err)
				return
			}
			draw_it(screen, ((size/512)*100)-float64(empty_parts)/(float64(width)/86), int(math.Round(x)), width, val[4], "/dev/"+disk+"p"+val[0], val[3])
		}
		list_it(screen, width, height, i, val, x, disk)
	}
	draw_box(screen, 7, 24, 47, pline, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFF00")), "")
}

func getPartitions() [][]string {
	// Run the parted command
	cmd := exec.Command("sudo", "parted", "-s", "/dev/nvme0n1", "unit", "GB", "print") // Replace "/dev/sda" with the target disk
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}

	// Convert output to string
	outputStr := string(output)

	// Regular expression to extract partition information
	// Example format: " 1     1049MB  100GB   100GB  ext4"
	re := regexp.MustCompile(`^\s*(\d+)\s{2,15}(\S+)\s{2,15}(\S+)\s{2,15}(\S+)(?:\s{2,15}(\S*))?(?:\s{2,15}((?:[A-Za-z]*\s??)*))?(?:\s{2,}((?:\S+(?:, )?)*))?$`)
	lines := strings.Split(outputStr, "\n")
	x := [][]string{}
	for _, line := range lines {
		matches := re.FindAllStringSubmatch(line, -1)
		if len(matches) == 1 {
			vals := []string{matches[0][1], matches[0][2], matches[0][3], matches[0][4], strings.TrimSpace(matches[0][5]), strings.TrimSpace(matches[0][6]), strings.TrimSpace(matches[0][7])}
			// fmt.Println(vals)
			x = append(x, vals)
		}
	}
	// Iterate through matches and extract partition details
	/*for _, match := range matches {
		// match[0] is the full match, match[1] is the partition number, and match[2] to match[5] are the other columns
		vals := []string{match[1], match[2], match[3], match[4], match[5]}
		x = append(x, vals)
	}*/
	return x
}

func mounted_parts(disk string, parts [][]string) {
	cmd := exec.Command("df", "-h") // Replace "/dev/sda" with the target disk
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}

	// Convert output to string
	outputStr := string(output)

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "/dev/"+disk) {
			re := regexp.MustCompile(`^(\S+?) .* (\S+?)$`)
			matches := re.FindAllStringSubmatch(line, -1)
			for i, part := range parts {
				// fmt.Println(string(line), []string{matches[0][1], matches[0][2]}, "/dev/"+disk+"p"+part[0], matches[0][1])
				if "/dev/"+disk+"p"+part[0] == matches[0][1] {
					parts[i] = append(parts[i], strings.TrimSpace(matches[0][2]))
				}
			}
		}
	}

	for i, part := range parts {
		if len(part) <= 7 {
			parts[i] = append(parts[i], "")
		}
	}
}

func part_labels(disk string, parts [][]string) {
	cmd := exec.Command("blkid") // Replace "/dev/sda" with the target disk
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}

	// Convert output to string
	outputStr := string(output)

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, "LABEL") {
			re := regexp.MustCompile(`^(\S+):.*?[^T]LABEL="([^"]*)"`)
			matches := re.FindAllStringSubmatch(line, -1)
			for i, part := range parts {
				if "/dev/"+disk+"p"+part[0] == matches[0][1] {
					// fmt.Println(matches, "/dev/"+disk+"p"+part[0])
					parts[i] = append(parts[i], strings.TrimSpace(matches[0][2]))
				}
			}
		}
	}

	for i, part := range parts {
		if len(part) <= 8 {
			parts[i] = append(parts[i], "")
		}
	}
}

func get_used_p(part, tip string) (float64, error) { // `tip` is turkish for type

	if tip == "ntfs" {

		cmd1 := exec.Command("sudo", "ntfsinfo", "-m", "/dev/"+part)
		output1, err := cmd1.Output()
		if err != nil {
			log.Fatal(err)
		}

		// Run the second command (grep Free) with input from the first command
		cmd2 := exec.Command("grep", "Free")
		cmd2.Stdin = bytes.NewReader(output1)
		output2, err2 := cmd2.Output()
		if err2 != nil {
			log.Fatal(err2)
		}

		// Run the third command (awk) with input from the second command
		cmd3 := exec.Command("awk", "{match($4, /\\(([0-9]+(\\.[0-9]+)?)%\\)/, arr); print arr[1]}")
		cmd3.Stdin = bytes.NewReader(output2)
		finalOutput, err3 := cmd3.Output()
		if err3 != nil {
			log.Fatal(err3)
		}

		// Print the final output
		val, err4 := strconv.ParseFloat(strings.TrimSpace(string(finalOutput)), 64)

		if err4 != nil {
			log.Fatal(err4)
		}

		return math.Round((100-val)*100) / 100, nil

	} else if tip == "ext4" {

		// echo "scale=5; 100 - (" (sudo tune2fs -l /dev/nvme0n1p5 | grep "Free blocks" | awk '{ print $3 }') "/" (sudo tune2fs -l /dev/nvme0n1p5 | grep "Block count" | awk '{ print $3 }') ") * 100" | bc
		cmd1 := exec.Command("sudo", "tune2fs", "-l", "/dev/"+part)
		output1, err := cmd1.Output()
		if err != nil {
			log.Fatal(err)
		}

		//fmt.Println(string(output1))

		lines := strings.Split(string(output1), "\n")

		needed := ""

		// Find the line starting with "Free blocks:"
		for _, line := range lines {
			if strings.HasPrefix(line, "Free blocks:") {

				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					number := strings.TrimSpace(parts[1]) // Remove any leading/trailing spaces
					needed = number                       // Print the number
				}
				break
			}
		}

		// Print the final output
		val_free, err4 := strconv.ParseFloat(strings.TrimSpace(string(needed)), 64)

		cmd4 := exec.Command("sudo", "tune2fs", "-l", "/dev/"+part)
		output4, err5 := cmd4.Output()
		if err5 != nil {
			log.Fatal(err5)
		}

		lines2 := strings.Split(string(output4), "\n")

		needed2 := ""

		// Find the line starting with "Free blocks:"
		for _, line := range lines2 {
			if strings.HasPrefix(line, "Block count:") {

				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					number := strings.TrimSpace(parts[1]) // Remove any leading/trailing spaces
					needed2 = number                      // Print the number
				}
				break
			}
		}

		val_total, err4 := strconv.ParseFloat(strings.TrimSpace(string(needed2)), 64)

		if err4 != nil {
			log.Fatal(err4)
		}

		//fmt.Println(val_free, val_total, val_free/val_total)

		return math.Round((100-(val_free/val_total)*100)*100) / 100, nil

	} else {
		return 0, errors.New("Cannot parse type: " + tip)
	}
}

func list_it(s tcell.Screen, width, _, index int, val []string, x float64, disk string) {
	start_n := 0
	used_p := "--"
	if x > 0 {
		used_p = fmt.Sprintf("%.2f%%", x)
	}
	colors := map[string]string{
		"ntfs":           "#61f8d0",
		"fat32":          "#76FF86",
		"ext4":           "#81AeCc",
		"linux-swap(v1)": "#E1968a",
	}
	tip := "unknown"
	if val[4] != "" {
		tip = val[4]
	}
	stuff := []string{
		"/dev/" + disk + "p" + val[0],
		val[5],
		tip,
		val[7],
		val[8],
		val[3],
		used_p,
		val[6],
	}
	categories_width := []int{
		15,
		20,
		15,
		10,
		14,
		8,
		8,
		10,
	}
	for i := 0; i < len(categories_width); i++ {
		start_n++
		width_n := int((float64(categories_width[i]) / 100) * float64(width))
		thing := " " + stuff[i]
		if i == 0 {
			// print_it(s, thing)
			print_at(s, start_n, index+10, thing, tcell.StyleDefault)
			if val[7] != "" {
				print_at(s, start_n, index+10, thing, tcell.StyleDefault)
				print_at(s, start_n+len(thing)+1, index+10, string(symbols[8]), tcell.StyleDefault.Foreground(tcell.GetColor("#D0D0D0")))
			} else {
				print_at(s, start_n, index+10, thing, tcell.StyleDefault)
			}
		} else if i == 2 {
			color, exists := colors[val[4]]
			var style tcell.Style
			if exists {
				style = tcell.StyleDefault.Foreground(tcell.GetColor(color))
			} else {
				style = tcell.StyleDefault.Foreground(tcell.GetColor("#888888"))
			}
			print_at(s, start_n+1, index+10, " ", style)
			print_at(s, start_n+2, index+10, thing, tcell.StyleDefault)
		} else {
			print_at(s, start_n, index+10, thing, tcell.StyleDefault)
		}
		start_n += width_n
	}
}

func print_head(s tcell.Screen, width, height int) {
	start_n := 0
	print_at(s, start_n, 7, "┌", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
	print_at(s, start_n, 8, "│", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
	print_at(s, start_n, height-2, "└", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
	for x := 9; x < height-2; x++ {
		print_at(s, start_n, x, "│", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
	}
	categories := []string{
		"Partition",
		"Name",
		"File System",
		"Mount Point",
		"Label",
		"Size",
		"Used %",
		"Flags",
	}
	categories_width := []int{
		15,
		20,
		15,
		10,
		14,
		8,
		8,
		10,
	}
	for i := 0; i < len(categories); i++ {
		width_n := int((float64(categories_width[i]) / 100) * float64(width))
		str := " " + categories[i]
		start_n++
		print_at(s, start_n, 7, strings.Repeat("─", width_n)+"┬", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
		print_at(s, start_n, height-2, strings.Repeat("─", width_n)+"┴", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
		print_at(s, start_n, 8, str, tcell.StyleDefault)
		print_at(s, start_n+width_n, 8, "│", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
		for x := 9; x < height-2; x++ {
			print_at(s, start_n+width_n, x, "│", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
		}
		print_at(s, start_n, 9, strings.Repeat("─", width_n)+"┼", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
		start_n += width_n
	}
	print_at(s, width-1, 7, "┐", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
	print_at(s, width-1, 8, "│", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
	print_at(s, width-1, height-2, "┘", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
	print_at(s, 0, 9, "├", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
	for x := 9; x < height-2; x++ {
		print_at(s, width-1, x, "│", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
	}
	print_at(s, width-1, 9, "┤", tcell.StyleDefault.Foreground(tcell.GetColor("#555555")))
}

func draw_it(s tcell.Screen, percent float64, filled, width int, tip, partn, size string) {
	chars_i := "🬂▐🬕🬲🬨🬷▌🬭"
	// chars_i := "─│╭╰╮╯"
	x := start + int((float64(percent)/100)*float64(width-2))
	if x >= width {
		x = width - 1
	} else if (x - start) < 2 {
		x = start + 1
	}
	colors := map[string]string{
		"ntfs":           "#41e8b0",
		"fat32":          "#46a046",
		"ext4":           "#314e6c",
		"linux-swap(v1)": "#c1665a",
	}
	color, exists := colors[tip]
	var style tcell.Style
	var style_box tcell.Style
	if exists {
		style = tcell.StyleDefault.Background(tcell.GetColor("#fce03f")).Foreground(tcell.GetColor(color))
		style_box = tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor(color))
	} else {
		style = tcell.StyleDefault.Background(tcell.GetColor("#fce03f")).Foreground(tcell.GetColor("#555555"))
		style_box = tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#555555"))
	}
	// print_it(s, tip)
	draw_box(s, start, 3, x, 6, style_box, chars_i)
	for y := 3; y <= 6; y++ {
		for x1 := start + 1; x1 < start+int((float64(filled)/100)*((float64(percent)/100)*float64(width))); x1++ {
			char, _, _, _ := s.GetContent(x1, y)
			if char == '\x00' {
				char = ' '
			}
			s.SetContent(x1, y, char, nil, style)
		}
		if filled > 0 {
			char, _, _, _ := s.GetContent(start, y)
			s.SetContent(start, y, char, nil, style)
		}
		if filled > 99 {
			char, _, _, _ := s.GetContent(x, y)
			s.SetContent(x, y, char, nil, style)
		}
		// s.SetContent(start+1, y, ' ', nil, style.Reverse(true))
		// s.SetContent(x-1, y, ' ', nil, style.Reverse(true))
	}
	if (x - start) > len(partn) {
		print_at(s, start+((x-start)/2)-(len(partn)/2), 4, partn, style.Foreground(tcell.GetColor("#000000")).Bold(true))
	}
	if (x - start) > len(size) {
		print_at(s, start+((x-start)/2)-(len(size)/2), 5, size, style.Foreground(tcell.GetColor("#000000")).Bold(true))
	}
	start = x + 2
}

func print_it(s tcell.Screen, str string) {
	s.SetContent(9, pline, symbols[7], nil, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFF00")))
	s.SetContent(10, pline, ' ', nil, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFFFF")))
	for i, char := range []rune(str) {
		s.SetContent(12+i, pline, char, nil, tcell.StyleDefault.Background(tcell.GetColor("#00000000")).Foreground(tcell.GetColor("#FFFFFF")))
	}
	pline++
}

func print_at(s tcell.Screen, x, y int, str string, style tcell.Style) {
	for i, char := range []rune(str) {
		s.SetContent(x+i, y, char, nil, style)
	}
}

func draw_box(s tcell.Screen, x1, y1, x2, y2 int, style tcell.Style, chars_i string) {
	if chars_i == "" {
		chars_i = "─│╭╰╮╯" // "━┃┏┗┓┛"
	}
	// print_it(s, fmt.Sprintf("Result: %d", utf8.RuneCountInString(chars_i)), style)
	if utf8.RuneCountInString(chars_i) == 6 {
		chars := []rune(chars_i)
		for x := x1; x <= x2; x++ {
			s.SetContent(x, y1, chars[0], nil, style)
		}
		for x := x1; x <= x2; x++ {
			s.SetContent(x, y2, chars[0], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			s.SetContent(x1, x, chars[1], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			s.SetContent(x2, x, chars[1], nil, style)
		}
		s.SetContent(x1, y1, chars[2], nil, style)
		s.SetContent(x1, y2, chars[3], nil, style)
		s.SetContent(x2, y1, chars[4], nil, style)
		s.SetContent(x2, y2, chars[5], nil, style)
	} else {
		chars := []rune(chars_i)
		for x := x1; x <= x2; x++ {
			s.SetContent(x, y1, chars[0], nil, style)
		}
		for x := x1; x <= x2; x++ {
			s.SetContent(x, y2, chars[7], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			s.SetContent(x1, x, chars[6], nil, style)
		}
		for x := y1 + 1; x < y2; x++ {
			s.SetContent(x2, x, chars[1], nil, style)
		}
		s.SetContent(x1, y1, chars[2], nil, style)
		s.SetContent(x1, y2, chars[3], nil, style)
		s.SetContent(x2, y1, chars[4], nil, style)
		s.SetContent(x2, y2, chars[5], nil, style)
	}
}
