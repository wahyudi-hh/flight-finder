package airport

import "flight-finder/internal/domain"
var airport = map[string]domain.Airport {
	"CGK": {
		Code: "CGK",
		City: "Jakarta",
	},
	"DPS": {
		Code: "DPS",
		City: "Denpasar",
	},
	"SUB": {
		Code: "SUB",
		City: "Surabaya",
	},
}

func Lookup(code string) (domain.Airport, bool) {
	airport, ok := airport[code]
	return airport, ok
}