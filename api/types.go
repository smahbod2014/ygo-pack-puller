package api

type Rarity string
type Foil string

const (
	RarityUltraRare Rarity = "Ultra Rare"
	RaritySuperRare Rarity = "Super Rare"
	RarityRare      Rarity = "Rare"
	RarityCommon    Rarity = "Common"
)

const (
	FoilNone   Foil = "normal"
	FoilGlossy Foil = "glossy"
	FoilRoyal  Foil = "ROYAL"
)

type SecretPackCardVariation struct {
	CardRarity Rarity `json:"card_rarity"`
}

type SecretPackCard struct {
	CardID         int                       `json:"card_id"`
	CardName       string                    `json:"card_name"`
	CardImg        string                    `json:"card_img"`
	CardVariations []SecretPackCardVariation `json:"card_variations"`
}

type MasterPackCard struct {
	Name   string `json:"name"`
	Pack   string `json:"pack"`
	Rarity Rarity `json:"rarity"`
}

type MasterPackQuery struct {
	Data []MasterPackCard `json:"data"`
}

type PulledCard struct {
	Rarity Rarity
	Foil   Foil
}

type ResultCard struct {
	CardID     int    `json:"card_id"`
	CardName   string `json:"card_name"`
	CardImg    string `json:"card_img"`
	CardRarity Rarity `json:"card_rarity"`
	CardFoil   Foil   `json:"card_foil"`
}
