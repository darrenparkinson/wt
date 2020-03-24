package cmd

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/darrenparkinson/wt/internal/webex"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	csvoutput bool
	listyear  int
	username  string
	password  string
	tenant    string
	siteID    string
	userid    string
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List WebEx Recordings",
	Long: `
The List command enables IT Administrators to list WebEx Recordings.`,
	Run: func(cmd *cobra.Command, args []string) {
		if tenant == "" || siteID == "" {
			log.Fatalln("tenant name and site ID required")
		}
		if username == "" || password == "" {
			username, password = credentials()
		}
		var list []webex.ListingRecording
		var err error
		if userid != "" {
			list, err = listRecordingsForUser(username, password, userid, tenant, siteID, listyear)
			if err != nil {
				log.Println(err)
			}
		} else {
			list, err = listRecordingsForYear(username, password, tenant, siteID, listyear)
			if err != nil {
				log.Println(err)
			}
		}
		if csvoutput {
			var headers []string
			if len(list) > 0 {
				headers = list[0].GetHeaders()
			}
			w := csv.NewWriter(os.Stdout)
			if err := w.Write(headers); err != nil {
				log.Fatalln("error writing record to csv:", err)
			}
			for _, r := range list {
				if err := w.Write(r.ToSlice()); err != nil {
					log.Fatalln("error writing record to csv:", err)
				}
			}
			w.Flush()
			if err := w.Error(); err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Printf("%12s\t%11s\t%12s\t%19s\t%s\n", "Host WebExID", "RecordingID", "Size", "CreateTime", "Name")
			fmt.Printf("%12s\t%11s\t%12s\t%19s\t%s\n", "------------", "-----------", "----", "----------", "----")
			for _, r := range list {
				fmt.Printf("%12s\t%11s\t%9.3f\t%19s\t%s\n", r.HostWebExID, r.RecordingID, r.RecordingSize, r.RecordingCreateTime, r.RecordingName)
			}
		}
	},
}

func listRecordingsForUser(username, password, userid, tenant, siteID string, year int) ([]webex.ListingRecording, error) {
	fromDate := time.Date(year, time.January, 01, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, time.December, 31, 23, 59, 59, 59, time.UTC)
	var allRecsForYear []webex.ListingRecording

	for currentDate := fromDate; currentDate.Before(endDate); currentDate = currentDate.AddDate(0, 0, 27) {
		toDate := currentDate.AddDate(0, 0, 27)
		if toDate.After(endDate) {
			toDate = endDate
		}
		recs, err := webex.GetRecordingListForUser(username, password, userid, tenant, siteID, currentDate, toDate)
		if err != nil {
			log.Fatalln(err)
		}
		for _, r := range recs.ListingBody.Recordings {
			allRecsForYear = append(allRecsForYear, r)
		}
	}
	return allRecsForYear, nil
}

func listRecordingsForYear(username, password, tenant, siteID string, year int) ([]webex.ListingRecording, error) {
	fromDate := time.Date(year, time.January, 01, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, time.December, 31, 23, 59, 59, 59, time.UTC)
	var allRecsForYear []webex.ListingRecording

	for currentDate := fromDate; currentDate.Before(endDate); currentDate = currentDate.AddDate(0, 0, 27) {
		toDate := currentDate.AddDate(0, 0, 27)
		if toDate.After(endDate) {
			toDate = endDate
		}
		recs, err := webex.GetRecordingList(username, password, tenant, siteID, currentDate, toDate)
		if err != nil {
			log.Fatalln(err)
		}
		for _, r := range recs.ListingBody.Recordings {
			allRecsForYear = append(allRecsForYear, r)
		}
	}
	return allRecsForYear, nil

}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&csvoutput, "csv", "c", false, "output CSV format instead of the default table format")
	listCmd.Flags().IntVarP(&listyear, "year", "y", 2020, "year to list")
	listCmd.Flags().StringVarP(&username, "username", "", "", "admin username, i.e. bjones")
	listCmd.Flags().StringVarP(&password, "password", "", "", "admin password")
	listCmd.Flags().StringVarP(&tenant, "tenant", "", "", "tenant name of your webex tenant without .webex.com, e.g. acme")
	listCmd.Flags().StringVarP(&siteID, "site", "", "", "site ID of your webex tenant, typically a six digit number")
	listCmd.Flags().StringVarP(&userid, "userid", "u", "", "userid of user recordings you wish to return for the given year")

}

func credentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("") // otherwise the output is on the password line.
	password := string(bytePassword)

	return strings.TrimSpace(username), password
}
