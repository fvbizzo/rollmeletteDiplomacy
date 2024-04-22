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
		Bases:  3,
	}

	England := Team{
		Name:   "England",
		Player: En,
		Armies: make(map[string]*Unit),
		Bases:  3,
	}
	France := Team{
		Name:   "France",
		Player: Fr,
		Bases:  3,
	}
	Germany := Team{
		Name:   "Germany",
		Player: Gr,
		Bases:  3,
	}
	Italy := Team{
		Name:   "Italy",
		Player: It,
		Bases:  3,
	}
	Russia := Team{
		Name:   "Russia",
		Player: Ru,
		Bases:  4,
	}
	Turkey := Team{
		Name:   "Turkey",
		Player: Tr,
		Armies: make(map[string]*Unit),
		Bases:  3,
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

func initializeRegions() map[string]*Region {
	Map := make(map[string]*Region)

	regionNames := []string{
		"Paris", "Burgundy", "English Channel", "London", "Liverpool", "Endinburgh", "Brest", "Marseilles", "Berlin", "Munich", "Kiel", "Rome", "Naples", "Venice",
		"Vienna", "Budapest", "Trieste", "Moscow", "St Petersburg", "Warsaw", "Constantinople", "Ankara", "Smyrna", "Belgium", "Holland", "Spain", "Portugal", "Denmark",
		"Sweden", "Norway", "Greece", "Serbia", "Bulgaria", "Rumania", "Tunis", "North Sea", "Irish Sea", "Mid Atlantic Ocean", "North Atlantic Ocean", "Norwegian Sea",
		"Skagerrak", "Baltic Sea", "Gulf of Bothnia", "Heligoland Bight", "Gulf of Lyons", "Tyrrhenian Sea", "Ionian Sea", "Aegean Sea", "Eastern Mediterranean",
		"Western Mediterranean", "Black Sea", "Adriatic Sea", "Barents Sea", "Sevastopol", "Apulia", "Armenia", "Bohemia", "Clyde", "Finland", "Galicia", "Gascony",
		"Livonia", "North Africa", "Picardy", "Piedmont", "Prussia", "Rhur", "Silesia", "Syria", "Tuscany", "Tyrolia", "Ukraine", "Wales", "Yorkshire", "Edinburgh", "Albania",
	}
	for _, name := range regionNames {
		Map[name] = &Region{Name: name, Occupied: false, Owner: "Neutral", SupplyCenter: false, Coastal: false, Sea: false}
	}
	connections := map[string][]string{
		"Paris":                 {"Burgundy", "Gascony", "Bugurndy", "Picardy"},
		"London":                {"North Sea", "English Channel", "Wales", "Yorkshire"},
		"Liverpool":             {"Irish Sea", "Edinburgh", "Yorkshire", "Wales", "Clyde", "North Atlantic Ocean"},
		"Yorkshire":             {"London", "Liverpool", "Edinburgh", "North Sea", "Wales"},
		"Edinburgh":             {"North Sea", "Norwegian Sea", "Liverpool", "Yorkshire", "Clyde"},
		"Brest":                 {"English Channel", "Mid Atlantic Ocean", "Gascony", "Paris", "Picardy"},
		"Marseilles":            {"Gulf of Lyon", "Burgundy", "Gascony", "Piedmont", "Spain"},
		"Burgundy":              {"Paris", "marseilles", "Gascony", "Picardy", "Belgium", "Rhur", "Munich"},
		"Picardy":               {"Paris", "Burgundy", "Belgium", "Brest", "English Channel"},
		"Gascony":               {"Spain", "Marseilles", "Burgundy", "Brest", "Mid Atlantic Ocean", "Paris"},
		"Munich":                {"Tyrolia", "Bohemia", "Burgundy", "Kiel", "Silesia", "Berlin", "Rhur"},
		"Berlin":                {"Baltic Sea", "Prussia", "Silesia", "Kiel", "Munich"},
		"Kiel":                  {"Helgoland Bight", "Baltic Sea", "Munich", "Berlin", "Rhur", "Holland"},
		"Prussia":               {"Berlin", "Silesia", "Warsaw", "Livonia", "Baltic Sea"},
		"Silesia":               {"Berlin", "Munich", "Warsaw", "Galicia", "Bohemia", "Prussia"},
		"Ruhr":                  {"Burgundy", "Belgium", "Holland", "Kiel", "Munich"},
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
		"St Petersburg":         {"Moscow", "Livonia", "Bothnia", "Finland", "barents sea", "Norway"},
		"Moscow":                {"St Petersburg", "Livonia", "Ukraine", "Sevastopol", "Warsaw"},
		"Livonia":               {"Moscow", "St Petersburg", "Prussia", "Baltic Sea", "Bothnia", "Warsaw"},
		"Warsaw":                {"Prussia", "Silesia", "Galicia", "Ukraine", "Moscow", "Livonia"},
		"Ukraine":               {"Moscow", "Warsaw", "Galicia", "Rumania", "Sevastopol"},
		"Sevastopol":            {"Rumania", "Moscow", "Ukraine", "Armenia", "Black Sea"},
		"Finland":               {"Sweden", "Norway", "St Petersburg", "Bothnia"},
		"Constantinople":        {"Bulgaria", "Aegean Sea", "Black Sea", "Smyrna", "Ankara"},
		"Ankara":                {"Constantinople", "Black Sea", "Smyrna", "Armenia"},
		"Smyrna":                {"Constantinople", "Aegean Sea", "Eastern Mediterranean", "Syria", "Ankara", "Armenia"},
		"Armenia":               {"Black Sea", "Sevastopol", "Ankara", "Syria", "Smyrna"},
		"Syria":                 {"Eastern Mediterranean", "Armenia", "Smyrna"},
		"Portugal":              {"Spain", "Mid Atlantic Ocean"},
		"Spain":                 {"Portugal", "Mid Atlantic Ocean", "Marseilles", "Gascony", "Gulf of Lyon", "Western Mediterranean"},
		"North Africa":          {"Mid Atlantic Ocean", "Western Mediterranean", "Tunis"},
		"Tunis":                 {"North Africa", "Western Mediterranean", "Thyrrean Sea", "Ionian Sea"},
		"Belgium":               {"Picardy", "Burgundy", "Rhur", "Holland", "English Channel", "North Sea"},
		"Holland":               {"Belgium", "North Sea", "Heligoland Bight", "Kiel", "Rhur"},
		"Denmark":               {"Kiel", "Heligoland Bight", "North Sea", "Skagerrak", "Sweden", "Baltic Sea"},
		"Norway":                {"North Sea", "Norwegian Sea", "Sweden", "Finland", "St Petersburg", "Barents Sea"},
		"Sweden":                {"Norway", "Baltic Sea", "Bothnia", "Finland", "Denmark", "Skagerrak"},
		"Serbia":                {"Budapest", "Rumania", "Bulgaria", "Albania", "Greece", "Trieste"},
		"Rumania":               {"Budapest", "Ukraine", "Sevastopol", "Bulgaria", "Black Sea", "Serbia"},
		"Bulgaria":              {"Rumania", "Black Sea", "Constantinople", "Aegean Sea", "Greece", "Serbia"},
		"Albania":               {"Trieste", "Serbia", "Greece", "Adriatic Sea", "Ionian Sea"},
		"Greece":                {"Albania", "Serbia", "Bulgaria", "Aegean Sea", "Ionian Sea"},
		"Black Sea":             {"Sevastopol", "Armenia", "Ankara", "Constantinople", "Bulgaria", "Rumania", "Aegean Sea"},
		"Aegean Sea":            {"Black Sea", "Eastern Mediterranean", "Ionian Sea", "Smyrna", "Constantinople", "Bulgaria", "Greece"},
		"Eatern Mediterranean":  {"Aegean Sea", "Ionian Sea", "Smyrna", "Syria"},
		"Ionian Sea":            {"Aegean Sea", "Eastern Mediterranean", "Adriatic Sea", "Thyrrean Sea", "Greece", "Albania", "Apulia", "Naples", "Tunis"},
		"Adriatic Sea":          {"Ionian Sea", "Venice", "Trieste", "Albania", "Apulia", "Rome"},
		"Thyrrean Sea":          {"Ionian Sea", "Western Mediterranean", "Rome", "Naples", "Tuscany", "Tunis"},
		"Gulf of Lyon":          {"Western Mediterranean", "Thyrrean Sea", "Tuscany", "Marseilles", "Piedmont", "Spain"},
		"Western Mediterranean": {"Gulf of Lyon", "Thyrrean Sea", "Mid Atlantic Ocean", "Spain", "North Africa", "Tunis"},
		"Mid Atlantic Ocean":    {"Western Mediterranean", "Portugal", "Spain", "Gascony", "Brest", "English Channel", "Irish Sea", "North Atlantic Ocean"},
		"North Atlantic Ocean":  {"Mid Atlantic Ocean", "Norwegian Sea", "Clyde", "Irish Sea", "liverpool"},
		"Irish Sea":             {"Mid Atlantic Ocean", "North Atlantic Sea", "Wales", "English Channel", "Liverpool"},
		"English Channel":       {"Mid Atlantic Ocean", "Brest", "Picardy", "London", "Wales", "North Sea", "Belgium"},
		"North Sea":             {"English Channel", "Heligoland Bight", "Skagerrak", "Norwegian Sea", "Belgium", "Holland", "Denmark", "Norway", "Edinburgh", "Yorkshire", "London"},
		"Heligoland Bight":      {"North Sea", "Holland", "Kiel", "Denmark"},
		"Skagerrak":             {"North Sea", "Norway", "Sweden", "Denmark"},
		"Baltic Sea":            {"Denmark", "Sweden", "Bothnia", "Livonia", "Prussia", "Berlin", "Kiel"},
		"Bothnia":               {"Sweden", "Finland", "St Petersburg", "Livonia", "Baltic Sea"},
		"Norwegian Sea":         {"North Atlantic Ocean", "Norway", "Barents Sea", "North Sea"},
		"Barents Sea":           {"Norwegian Sea", "St Petersburg", "Norway"},
	}

	for name, neighbors := range connections {
		for _, neighbor := range neighbors {
			Map[name].Neighbors = append(Map[name].Neighbors, Map[neighbor])
		}
	}
	Map["London"].SupplyCenter = true

	return Map
}
