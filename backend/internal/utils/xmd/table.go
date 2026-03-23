package xmd

import "strings"

func CreateMarkdownTable(head []string, body [][]string) string {
	separator := "|"
	columnWidths := make([]int, len(head))

	//Calculate maximum column width based on headers
	for i, val := range head {
		columnWidths[i] = len(val)
	}

	//Calculate maximum column widths based on body data
	for _, row := range body {
		for i, val := range row {
			if i < len(head) && len(val) > columnWidths[i] {
				columnWidths[i] = len(val)
			}
		}
	}

	var str strings.Builder

	//Build header row
	for i, val := range head {
		str.WriteString(separator + " " + padString(val, columnWidths[i]) + " ")
	}
	str.WriteString(separator + "\n")

	//Build separator row
	for _, width := range columnWidths {
		str.WriteString(separator + " " + strings.Repeat("-", width) + " ")
	}
	str.WriteString(separator + "\n")

	//Build body rows
	//Ensure we only loop through columns present in the header
	for _, row := range body {
		for i := 0; i < len(head); i++ {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			str.WriteString(separator + " " + padString(val, columnWidths[i]) + " ")
		}
		str.WriteString(separator + "\n")
	}

	return str.String()
}

// Helper function to pad a string to the required length
func padString(val string, length int) string {
	return val + strings.Repeat(" ", length-len(val))
}
