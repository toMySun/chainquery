package db

import (
	"database/sql"

	g "github.com/lbryio/chainquery/swagger/clients/goclient"

	"github.com/sirupsen/logrus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries"
)

// AddressSummary summarizes information for an address from chainquery database
type AddressSummary struct {
	ID            uint64  `boil:"id"`
	Address       string  `boil:"address"`
	TotalReceived float64 `boil:"total_received"`
	TotalSent     float64 `boil:"total_sent"`
	Balance       float64 `boil:"balance"`
}

// GetTableStatus provides size information for the tables in the chainquery database
func GetTableStatus() (*g.TableStatus, error) {
	println("here2")
	stats := g.TableStatus{}
	rows, err := boil.GetDB().Query(
		`SELECT TABLE_NAME as "table",` +
			`SUM(TABLE_ROWS) as "rows" ` +
			`FROM INFORMATION_SCHEMA.TABLES ` +
			`WHERE TABLE_SCHEMA = "lbrycrd" ` +
			`GROUP BY TABLE_NAME;`)

	if err != nil {
		return nil, err
	}
	defer closeRows(rows)
	var statrows []g.TableSize
	for rows.Next() {
		var stat g.TableSize
		err = rows.Scan(&stat.TableName, &stat.NrRows)
		if err != nil {
			return nil, err
		}
		statrows = append(statrows, stat)
	}

	stats.Status = statrows

	return &stats, nil
}

// GetAddressSummary returns summary information of an address in the chainquery database.
func GetAddressSummary(address string) (*AddressSummary, error) {
	addressSummary := AddressSummary{}
	err := queries.RawG(
		`SELECT address.address, `+
			`SUM(ta.credit_amount) AS total_received, `+
			`SUM(ta.debit_amount) AS total_sent,`+
			`(SUM(ta.credit_amount) - SUM(ta.debit_amount)) AS balance `+
			`FROM address LEFT JOIN transaction_address as ta ON ta.address_id = address.id `+
			`WHERE address.address=? `+
			`GROUP BY address.address `, address).Bind(&addressSummary)

	if err != nil {
		return nil, err
	}

	return &addressSummary, nil

}

func closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		logrus.Error("Closing rows error: ", err)
	}
}
