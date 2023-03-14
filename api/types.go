package api

type Rarity string
type Foil string

const (
	RarityUltraRare Rarity = "UR"
	RaritySuperRare Rarity = "SR"
	RarityRare      Rarity = "R"
	RarityCommon    Rarity = "N"
)

const (
	FoilNone   Foil = "normal"
	FoilGlossy Foil = "glossy"
	FoilRoyal  Foil = "royal"
)

type PulledCard struct {
	Rarity Rarity
	Foil   Foil
}

type ResultCard struct {
	CardID     string `json:"card_id"`
	CardName   string `json:"card_name"`
	CardImg    string `json:"card_img"`
	CardRarity Rarity `json:"card_rarity"`
	CardFoil   Foil   `json:"card_foil"`
}

type MDMPackType string

const (
	MDMPackTypeSecretPack    MDMPackType = "Secret Pack"
	MDMPackTypeSelectionPack MDMPackType = "Selection Pack"
	MDMPackTypeMasterPack    MDMPackType = "Normal Pack"
)

type MDMPack struct {
	ID   string      `json:"_id"`
	Type MDMPackType `json:"type"`
	Name string      `json:"name"`
}

type MDMCard struct {
	ID       string `json:"_id"`
	KonamiID string `json:"konamiID"`
	Name     string `json:"name"`
	Rarity   Rarity `json:"rarity"`
}
