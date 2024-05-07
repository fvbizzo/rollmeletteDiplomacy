package main

import (
	"github.com/ethereum/go-ethereum/common"
)

func initializePlayers(
	Au common.Address,
	En common.Address,
	Fr common.Address,
	Gr common.Address,
	It common.Address,
	Ru common.Address,
	Tr common.Address,
) map[common.Address]*Team {

	Players := make(map[common.Address]*Team)

	Austria := Team{
		Name:   "Austria",
		Player: Au,
		Armies: map[int]string{
			1: "Vienna",
			2: "Budapest",
			3: "Triest",
		},
		Bases: 3,
	}

	England := Team{
		Name:   "England",
		Player: En,
		Armies: map[int]string{
			4: "London",
			5: "Liverpool",
			6: "Edinburgh",
		},
		Bases: 3,
	}
	France := Team{
		Name:   "France",
		Player: Fr,
		Armies: map[int]string{
			7: "Paris",
			8: "Brest",
			9: "Marseilles",
		},
		Bases: 3,
	}
	Germany := Team{
		Name:   "Germany",
		Player: Gr,
		Armies: map[int]string{
			10: "Berlin",
			11: "Munich",
			12: "Kiel",
		},
		Bases: 3,
	}
	Italy := Team{
		Name:   "Italy",
		Player: It,
		Armies: map[int]string{
			13: "Rome",
			14: "Venice",
			15: "Naples",
		},
		Bases: 3,
	}
	Russia := Team{
		Name:   "Russia",
		Player: Ru,
		Armies: map[int]string{
			16: "Moscow",
			17: "St Petersburg",
			18: "Warsaw",
			19: "Sevastopol",
		},
		Bases: 4,
	}
	Turkey := Team{
		Name:   "Turkey",
		Player: Tr,
		Armies: map[int]string{
			20: "Constantinople",
			21: "Smyrna",
			22: "Ankara",
		},
		Bases: 3,
	}

	Players[Au] = &Austria
	Players[En] = &England
	Players[Fr] = &France
	Players[Gr] = &Germany
	Players[It] = &Italy
	Players[Ru] = &Russia
	Players[Tr] = &Turkey

	return Players
}

// Function to initialize map at the start of the game
func initializeRegions() map[string]*Region {
	Map := make(map[string]*Region)

	//All Region names listed as strings
	regionNames := []string{
		"Paris", "Burgundy", "English Channel", "London", "Liverpool", "Endinburgh", "Brest", "Marseilles", "Berlin", "Munich", "Kiel", "Rome", "Naples", "Venice",
		"Vienna", "Budapest", "Trieste", "Moscow", "St Petersburg", "Warsaw", "Constantinople", "Ankara", "Smyrna", "Belgium", "Holland", "Spain", "Portugal", "Denmark",
		"Sweden", "Norway", "Greece", "Serbia", "Bulgaria", "Rumania", "Tunis", "North Sea", "Irish Sea", "Mid Atlantic Ocean", "North Atlantic Ocean", "Norwegian Sea",
		"Skagerrak", "Baltic Sea", "Gulf of Bothnia", "Heligoland Bight", "Gulf of Lyon", "Tyrrhenian Sea", "Ionian Sea", "Aegean Sea", "Eastern Mediterranean",
		"Western Mediterranean", "Black Sea", "Adriatic Sea", "Barents Sea", "Sevastopol", "Apulia", "Armenia", "Bohemia", "Clyde", "Finland", "Galicia", "Gascony",
		"Livonia", "North Africa", "Picardy", "Piedmont", "Prussia", "Rhur", "Silesia", "Syria", "Tuscany", "Tyrolia", "Ukraine", "Wales", "Yorkshire", "Edinburgh", "Albania",
	}

	//Regions are created as as owner neutral, not Coastal, not sea, not occupied  and not a supply center
	for _, name := range regionNames {
		Map[name] = &Region{Name: name, Occupied: false, Owner: "Neutral", SupplyCenter: false, Coastal: false, Sea: false}
	}

	//All the regions connections as a map of strings
	connections := map[string][]string{
		"Paris":                 {"Burgundy", "Gascony", "Burgundy", "Picardy", "English Channel", "Brest"},
		"London":                {"North Sea", "English Channel", "Wales", "Yorkshire"},
		"Liverpool":             {"Irish Sea", "Edinburgh", "Yorkshire", "Wales", "Clyde", "North Atlantic Ocean"},
		"Yorkshire":             {"London", "Liverpool", "Edinburgh", "North Sea", "Wales"},
		"Edinburgh":             {"North Sea", "Norwegian Sea", "Liverpool", "Yorkshire", "Clyde"},
		"Brest":                 {"English Channel", "Mid Atlantic Ocean", "Gascony", "Paris", "Picardy"},
		"Marseilles":            {"Gulf of Lyon", "Burgundy", "Gascony", "Piedmont", "Spain"},
		"Burgundy":              {"Paris", "Marseilles", "Gascony", "Picardy", "Belgium", "Rhur", "Munich"},
		"Picardy":               {"Paris", "Burgundy", "Belgium", "Brest", "English Channel"},
		"Gascony":               {"Spain", "Marseilles", "Burgundy", "Brest", "Mid Atlantic Ocean", "Paris"},
		"Munich":                {"Tyrolia", "Bohemia", "Burgundy", "Kiel", "Silesia", "Berlin", "Rhur"},
		"Berlin":                {"Baltic Sea", "Prussia", "Silesia", "Kiel", "Munich"},
		"Kiel":                  {"Heligoland Bight", "Baltic Sea", "Munich", "Berlin", "Rhur", "Holland"},
		"Prussia":               {"Berlin", "Silesia", "Warsaw", "Livonia", "Baltic Sea"},
		"Silesia":               {"Berlin", "Munich", "Warsaw", "Galicia", "Bohemia", "Prussia"},
		"Rhur":                  {"Burgundy", "Belgium", "Holland", "Kiel", "Munich"},
		"Rome":                  {"Tyrrhenian Sea", "Tuscany", "Venice", "Naples", "Apulia"},
		"Naples":                {"Tyrrhenian Sea", "Apulia", "Rome", "Ionian Sea"},
		"Venice":                {"Adriatic Sea", "Trieste", "Tyrolia", "Piedmont"},
		"Tuscany":               {"Rome", "Venice", "Piedmont", "Tyrrhenian Sea", "Gulf of Lyon"},
		"Piedmont":              {"Marseilles", "Gulf of Lyon", "Venice", "Tuscany"},
		"Apulia":                {"Naples", "Venice", "Ionian Sea", "Rome", "Adriatic Sea"},
		"Vienna":                {"Bohemia", "Galicia", "Budapest", "Trieste", "Tyrolia"},
		"Budapest":              {"Galicia", "Serbia", "Rumania", "Vienna", "Trieste"},
		"Trieste":               {"Adriatic Sea", "Venice", "Tyrolia", "Serbia", "Albania", "Vienna", "Budapest"},
		"Galicia":               {"Warsaw", "Silesia", "Budapest", "Vienna", "Ukraine", "Rumania", "Bohemia"},
		"Bohemia":               {"Munich", "Silesia", "Galicia", "Vienna", "Tyrolia"},
		"Tyrolia":               {"Munich", "Bohemia", "Venice", "Trieste", "Vienna", "Piedmont"},
		"St Petersburg":         {"Moscow", "Livonia", "Gulf of Bothnia", "Finland", "barents sea", "Norway"},
		"Moscow":                {"St Petersburg", "Livonia", "Ukraine", "Sevastopol", "Warsaw"},
		"Livonia":               {"Moscow", "St Petersburg", "Prussia", "Baltic Sea", "Gulf of Bothnia", "Warsaw"},
		"Warsaw":                {"Prussia", "Silesia", "Galicia", "Ukraine", "Moscow", "Livonia"},
		"Ukraine":               {"Moscow", "Warsaw", "Galicia", "Rumania", "Sevastopol"},
		"Sevastopol":            {"Rumania", "Moscow", "Ukraine", "Armenia", "Black Sea"},
		"Finland":               {"Sweden", "Norway", "St Petersburg", "Gulf of Bothnia"},
		"Constantinople":        {"Bulgaria", "Aegean Sea", "Black Sea", "Smyrna", "Ankara"},
		"Ankara":                {"Constantinople", "Black Sea", "Smyrna", "Armenia"},
		"Smyrna":                {"Constantinople", "Aegean Sea", "Eastern Mediterranean", "Syria", "Ankara", "Armenia"},
		"Armenia":               {"Black Sea", "Sevastopol", "Ankara", "Syria", "Smyrna"},
		"Syria":                 {"Eastern Mediterranean", "Armenia", "Smyrna"},
		"Portugal":              {"Spain", "Mid Atlantic Ocean"},
		"Spain":                 {"Portugal", "Mid Atlantic Ocean", "Marseilles", "Gascony", "Gulf of Lyon", "Western Mediterranean"},
		"North Africa":          {"Mid Atlantic Ocean", "Western Mediterranean", "Tunis"},
		"Tunis":                 {"North Africa", "Western Mediterranean", "Tyrrhenian Sea", "Ionian Sea"},
		"Belgium":               {"Picardy", "Burgundy", "Rhur", "Holland", "English Channel", "North Sea"},
		"Holland":               {"Belgium", "North Sea", "Heligoland Bight", "Kiel", "Rhur"},
		"Denmark":               {"Kiel", "Heligoland Bight", "North Sea", "Skagerrak", "Sweden", "Baltic Sea"},
		"Norway":                {"North Sea", "Norwegian Sea", "Sweden", "Finland", "St Petersburg", "Barents Sea"},
		"Sweden":                {"Norway", "Baltic Sea", "Gulf of Bothnia", "Finland", "Denmark", "Skagerrak"},
		"Serbia":                {"Budapest", "Rumania", "Bulgaria", "Albania", "Greece", "Trieste"},
		"Rumania":               {"Budapest", "Ukraine", "Sevastopol", "Bulgaria", "Black Sea", "Serbia"},
		"Bulgaria":              {"Rumania", "Black Sea", "Constantinople", "Aegean Sea", "Greece", "Serbia"},
		"Albania":               {"Trieste", "Serbia", "Greece", "Adriatic Sea", "Ionian Sea"},
		"Greece":                {"Albania", "Serbia", "Bulgaria", "Aegean Sea", "Ionian Sea"},
		"Black Sea":             {"Sevastopol", "Armenia", "Ankara", "Constantinople", "Bulgaria", "Rumania", "Aegean Sea"},
		"Aegean Sea":            {"Black Sea", "Eastern Mediterranean", "Ionian Sea", "Smyrna", "Constantinople", "Bulgaria", "Greece"},
		"Eastern Mediterranean": {"Aegean Sea", "Ionian Sea", "Smyrna", "Syria"},
		"Ionian Sea":            {"Aegean Sea", "Eastern Mediterranean", "Adriatic Sea", "Tyrrhenian Sea", "Greece", "Albania", "Apulia", "Naples", "Tunis"},
		"Adriatic Sea":          {"Ionian Sea", "Venice", "Trieste", "Albania", "Apulia", "Rome"},
		"Tyrrhenian Sea":        {"Ionian Sea", "Western Mediterranean", "Rome", "Naples", "Tuscany", "Tunis"},
		"Gulf of Lyon":          {"Western Mediterranean", "Tyrrhenian Sea", "Tuscany", "Marseilles", "Piedmont", "Spain"},
		"Western Mediterranean": {"Gulf of Lyon", "Tyrrhenian Sea", "Mid Atlantic Ocean", "Spain", "North Africa", "Tunis"},
		"Mid Atlantic Ocean":    {"Western Mediterranean", "Portugal", "Spain", "Gascony", "Brest", "English Channel", "Irish Sea", "North Atlantic Ocean"},
		"North Atlantic Ocean":  {"Mid Atlantic Ocean", "Norwegian Sea", "Clyde", "Irish Sea", "liverpool"},
		"Irish Sea":             {"Mid Atlantic Ocean", "North Atlantic Sea", "Wales", "English Channel", "Liverpool"},
		"English Channel":       {"Mid Atlantic Ocean", "Brest", "Picardy", "London", "Wales", "North Sea", "Belgium"},
		"North Sea":             {"English Channel", "Heligoland Bight", "Skagerrak", "Norwegian Sea", "Belgium", "Holland", "Denmark", "Norway", "Edinburgh", "Yorkshire", "London"},
		"Heligoland Bight":      {"North Sea", "Holland", "Kiel", "Denmark"},
		"Skagerrak":             {"North Sea", "Norway", "Sweden", "Denmark"},
		"Baltic Sea":            {"Denmark", "Sweden", "Gulf of Bothnia", "Livonia", "Prussia", "Berlin", "Kiel"},
		"Gulf of Bothnia":       {"Sweden", "Finland", "St Petersburg", "Livonia", "Baltic Sea"},
		"Norwegian Sea":         {"North Atlantic Ocean", "Norway", "Barents Sea", "North Sea"},
		"Barents Sea":           {"Norwegian Sea", "St Petersburg", "Norway"},
	}

	for name, neighbors := range connections {
		for _, neighbor := range neighbors {
			n := neighbor
			Map[name].Neighbors = append(Map[name].Neighbors, &n)
		}
	}

	//Adding the supplycenter bool to the supply regions
	supplyCenters := []string{
		"London", "Liverpool", "Edinburgh",
		"Paris", "Brest", "Marseilles",
		"Kiel", "Berlin", "Munich",
		"Rome", "Venice", "Naples",
		"Vienna", "Trieste", "Budapest",
		"Constantinople", "Smyrna", "Ankara",
		"Moscow", "St Petersburg", "Sevastopol", "Warsaw",
		"Portugal", "Spain", "Tunis",
		"Norway", "Sweden", "Denmark",
		"Rumania", "Serbia", "Bulgaria",
		"Belgium", "Holland", "Greece",
	}

	for _, name := range supplyCenters {
		Map[name].SupplyCenter = true
	}

	//Giving an owner to all the starting position regions for each player
	Map["Budapest"].Owner = "Austria"
	Map["Vienna"].Owner = "Austria"
	Map["Trieste"].Owner = "Austria"
	Map["Galicia"].Owner = "Austria"
	Map["Bohemia"].Owner = "Austria"
	Map["Tyrolia"].Owner = "Austria"

	Map["London"].Owner = "England"
	Map["Yorkshire"].Owner = "England"
	Map["Edinburgh"].Owner = "England"
	Map["Liverpool"].Owner = "England"
	Map["Clyde"].Owner = "England"
	Map["Wales"].Owner = "England"

	Map["Paris"].Owner = "France"
	Map["Burgundy"].Owner = "France"
	Map["Brest"].Owner = "France"
	Map["Marseilles"].Owner = "France"
	Map["Gascony"].Owner = "France"
	Map["Picardy"].Owner = "France"

	Map["Berlin"].Owner = "Germany"
	Map["Kiel"].Owner = "Germany"
	Map["Munich"].Owner = "Germany"
	Map["Prussia"].Owner = "Germany"
	Map["Silesia"].Owner = "Germany"
	Map["Rhur"].Owner = "Germany"

	Map["Rome"].Owner = "Italy"
	Map["Venice"].Owner = "Italy"
	Map["Apulia"].Owner = "Italy"
	Map["Naples"].Owner = "Italy"
	Map["Tuscany"].Owner = "Italy"
	Map["Piedmont"].Owner = "Italy"

	Map["Moscow"].Owner = "Russia"
	Map["St Petersburg"].Owner = "Russia"
	Map["Sevastopol"].Owner = "Russia"
	Map["Livonia"].Owner = "Russia"
	Map["Warsaw"].Owner = "Russia"
	Map["Ukraine"].Owner = "Russia"
	Map["Finland"].Owner = "Russia"

	Map["Constantinople"].Owner = "Turkey"
	Map["Smyrna"].Owner = "Turkey"
	Map["Ankara"].Owner = "Turkey"
	Map["Armenia"].Owner = "Turkey"
	Map["Syria"].Owner = "Turkey"

	//Setting Sea and coastal regions

	seaReagions := []string{
		"Black Sea", "Aegean Sea", "Eastern Mediterranean", "Ionian Sea", "Adriatic Sea",
		"Tyrrhenian Sea", "Gulf of Lyon", "Western Mediterranean", "Mid Atlantic Ocean",
		"North Atlantic Ocean", "Irish Sea", "English Channel", "North Sea", "Barents Sea",
		"Heligoland Bight", "Skagerrak", "Baltic Sea", "Gulf of Bothnia", "Norwegian Sea",
	}

	for _, name := range seaReagions {
		Map[name].Sea = true
	}

	coastalRegions := []string{
		"Clyde", "London", "Edinburgh", "Liverpool", "Wales", "Yorkshire", "Brest", "Picardy",
		"Gascony", "Marseilles", "Kiel", "Berlin", "Prussia", "Rome", "Naples", "Venice",
		"Tuscany", "Piedmont", "Apulia", "Trieste", "Finland", "St Petersburg", "Sevastopol",
		"Livonia", "Constantinople", "Smyrna", "Ankara", "Syria", "Armenia", "Portugal", "Spain",
		"North Africa", "Tunis", "Belgium", "Holland", "Denmark", "Sweden", "Norway", "Albania",
		"Greece", "Bulgaria", "Rumania",
	}

	for _, name := range coastalRegions {
		Map[name].Coastal = true
	}

	//Setting the initial army positions for each team

	Map["Budapest"].Occupied = true
	Map["Vienna"].Occupied = true
	Map["Trieste"].Occupied = true

	Map["London"].Occupied = true
	Map["Edinburgh"].Occupied = true
	Map["Liverpool"].Occupied = true

	Map["Paris"].Occupied = true
	Map["Brest"].Occupied = true
	Map["Marseilles"].Occupied = true

	Map["Berlin"].Occupied = true
	Map["Kiel"].Occupied = true
	Map["Munich"].Occupied = true

	Map["Rome"].Occupied = true
	Map["Venice"].Occupied = true
	Map["Naples"].Occupied = true

	Map["Moscow"].Occupied = true
	Map["St Petersburg"].Occupied = true
	Map["Sevastopol"].Occupied = true
	Map["Warsaw"].Occupied = true

	Map["Constantinople"].Occupied = true
	Map["Smyrna"].Occupied = true
	Map["Ankara"].Occupied = true

	return Map
}

func initializeUnits() map[int]*Unit {
	Units := map[int]*Unit{

		1:  {ID: 1, Type: "army", Position: "Vienna", Owner: "Austria", CurrentOrder: Orders{UnitID: 1, Ordertype: "hold"}, Retreating: false},
		2:  {ID: 2, Type: "army", Position: "Budapest", Owner: "Austria", CurrentOrder: Orders{UnitID: 2, Ordertype: "hold"}, Retreating: false},
		3:  {ID: 3, Type: "navy", Position: "Triest", Owner: "Austria", CurrentOrder: Orders{UnitID: 3, Ordertype: "hold"}, Retreating: false},
		4:  {ID: 4, Type: "navy", Position: "London", Owner: "England", CurrentOrder: Orders{UnitID: 4, Ordertype: "hold"}, Retreating: false},
		5:  {ID: 5, Type: "army", Position: "Liverpool", Owner: "England", CurrentOrder: Orders{UnitID: 5, Ordertype: "hold"}, Retreating: false},
		6:  {ID: 6, Type: "navy", Position: "Edinburgh", Owner: "England", CurrentOrder: Orders{UnitID: 6, Ordertype: "hold"}, Retreating: false},
		7:  {ID: 7, Type: "army", Position: "Paris", Owner: "France", CurrentOrder: Orders{UnitID: 7, Ordertype: "hold"}, Retreating: false},
		8:  {ID: 8, Type: "navy", Position: "Brest", Owner: "France", CurrentOrder: Orders{UnitID: 8, Ordertype: "hold"}, Retreating: false},
		9:  {ID: 9, Type: "army", Position: "Marseilles", Owner: "France", CurrentOrder: Orders{UnitID: 9, Ordertype: "hold"}, Retreating: false},
		10: {ID: 10, Type: "army", Position: "Berlin", Owner: "Germany", CurrentOrder: Orders{UnitID: 10, Ordertype: "hold"}, Retreating: false},
		11: {ID: 11, Type: "army", Position: "Munich", Owner: "Germany", CurrentOrder: Orders{UnitID: 11, Ordertype: "hold"}, Retreating: false},
		12: {ID: 12, Type: "navy", Position: "Kiel", Owner: "Germany", CurrentOrder: Orders{UnitID: 12, Ordertype: "hold"}, Retreating: false},
		13: {ID: 13, Type: "army", Position: "Rome", Owner: "Italy", CurrentOrder: Orders{UnitID: 13, Ordertype: "hold"}, Retreating: false},
		14: {ID: 14, Type: "army", Position: "Venice", Owner: "Italy", CurrentOrder: Orders{UnitID: 14, Ordertype: "hold"}, Retreating: false},
		15: {ID: 15, Type: "navy", Position: "Naples", Owner: "Italy", CurrentOrder: Orders{UnitID: 15, Ordertype: "hold"}, Retreating: false},
		16: {ID: 16, Type: "Army", Position: "Moscow", Owner: "Russia", CurrentOrder: Orders{UnitID: 16, Ordertype: "hold"}, Retreating: false},
		17: {ID: 17, Type: "navy", Position: "St Petersburg", Owner: "Russia", CurrentOrder: Orders{UnitID: 17, Ordertype: "hold"}, Retreating: false},
		18: {ID: 18, Type: "army", Position: "Warsaw", Owner: "Russia", CurrentOrder: Orders{UnitID: 18, Ordertype: "hold"}, Retreating: false},
		19: {ID: 19, Type: "navy", Position: "Sevastopol", Owner: "Russia", CurrentOrder: Orders{UnitID: 19, Ordertype: "hold"}, Retreating: false},
		20: {ID: 20, Type: "army", Position: "Constantinople", Owner: "Turkey", CurrentOrder: Orders{UnitID: 20, Ordertype: "hold"}, Retreating: false},
		21: {ID: 21, Type: "army", Position: "Smyrna", Owner: "Turkey", CurrentOrder: Orders{UnitID: 21, Ordertype: "hold"}, Retreating: false},
		22: {ID: 22, Type: "navy", Position: "Ankara", Owner: "Turkey", CurrentOrder: Orders{UnitID: 22, Ordertype: "hold"}, Retreating: false},
	}
	return Units
}
