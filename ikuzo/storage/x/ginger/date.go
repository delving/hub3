package ginger

import (
	"strings"

	"github.com/araddon/dateparse"
)

func reverseDates(date string) (string, error) {
	cleanDate := strings.ReplaceAll(date, "-", "/")
	t, err := dateparse.ParseLocal(cleanDate)
	if err != nil {
		return "", err
	}

	//fmt.Printf("%s", t.Format("2006-01-02"))
	return t.Format("2006-01-02"), nil
}
