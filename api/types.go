package api

import (
	"encoding/json"
	"strconv"
)

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
	ID       string  `json:"_id"`
	KonamiID FlexInt `json:"konamiID"`
	Name     string  `json:"name"`
	Rarity   Rarity  `json:"rarity"`
}

type FlexInt int

func (flexInt *FlexInt) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		return json.Unmarshal(b, (*int)(flexInt))
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*flexInt = FlexInt(i)
	return nil
}
