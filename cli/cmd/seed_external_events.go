package cmd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/dxe/adb/cli/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	_ "github.com/go-sql-driver/mysql"
)

// seedExternalEventPageID is the SF Bay Facebook page, which is the only page the external events admin page shows.
// Keep in sync with model.SFBayPageID (the server module isn't importable from the CLI module).
const seedExternalEventPageID = 1377014279263790

// seedExternalEvents lists the test events to insert, with the number of days from today that each one starts on and
// a Bay Area venue.
var seedExternalEvents = []struct {
	ID      int64
	Name    string
	DaysOut int
	Venue   string
	Address string
	City    string
	Zip     string
	Lat     float64
	Lng     float64
}{
	{2032052407456743, "Seed External Event A", 30, "Seed Venue Berkeley", "2140 Shattuck Ave", "Berkeley", "94704", 37.870200, -122.268400},
	{1080428957901856, "Seed External Event B", 31, "Seed Venue Oakland", "1 Frank H Ogawa Plaza", "Oakland", "94612", 37.805300, -122.272700},
	{1416265224018809, "Seed External Event C", 31, "Seed Venue San Francisco", "1 Dr Carlton B Goodlett Pl", "San Francisco", "94102", 37.779300, -122.419200},
	{1718100289270785, "Seed External Event D", 32, "Seed Venue San Jose", "200 E Santa Clara St", "San Jose", "95113", 37.337600, -121.887400},
}

func init() {
	seedCmd.AddCommand(seedExternalEventsCmd)
}

var seedExternalEventsCmd = &cobra.Command{
	Use:   "external-events",
	Short: "Seed 4 test Facebook events 30-32 days out (two on the same day), replacing them if they already exist",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireNotProd(); err != nil {
			return err
		}

		conn, err := sqlx.Connect("mysql", config.DBDataSource())
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer func() { _ = conn.Close() }()

		now := time.Now()
		for _, event := range seedExternalEvents {
			// 6pm, `DaysOut` days from today.
			start := time.Date(now.Year(), now.Month(), now.Day()+event.DaysOut, 18, 0, 0, 0, now.Location())
			if _, err := conn.Exec(
				`REPLACE INTO fb_events (id, page_id, name, description, start_time, end_time,
					location_name, location_city, location_state, location_country, location_address, location_zip,
					lat, lng, cover, attending_count, interested_count)
				VALUES (?, ?, ?, 'Seeded test event.', ?, ?,
					?, ?, 'California', 'United States', ?, ?,
					?, ?, ?, 10, 20)`,
				event.ID, seedExternalEventPageID, event.Name,
				start.Format("2006-01-02 15:04:05"), start.Add(2*time.Hour).Format("2006-01-02 15:04:05"),
				event.Venue, event.City, event.Address, event.Zip, event.Lat, event.Lng,
				seedExternalEventCover(event.City),
			); err != nil {
				return fmt.Errorf("failed to insert external event %q: %w", event.Name, err)
			}
		}

		fmt.Printf("Seeded %d external events on page %d\n", len(seedExternalEvents), seedExternalEventPageID)
		return nil
	},
}

// seedExternalEventCover returns a placeholder cover image URL, standing in for the Facebook-hosted image that the
// event sync would normally store.
func seedExternalEventCover(city string) string {
	return "https://placehold.co/1200x630/png?text=" + url.QueryEscape(city)
}
