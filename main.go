package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

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

// func (card PulledCard) String() string {
// 	s := fmt.Sprintf("%s (%s) (%s)", card.Card.Name, card.Card.Rarity, card.Foil)
// 	if card.Card.Rarity == RaritySR {
// 		s = "+++ " + s + " +++"
// 	} else if card.Card.Rarity == RarityUR {
// 		s = "$$$$$$ " + s + " $$$$$$"
// 	}
// 	return s
// }

func main() {
	rand.Seed(time.Now().Unix())

	packName := "Rapid Aircraft Advancement"

	// Get the cards from the pack
	secretPackCardsResponse, err := http.Get(fmt.Sprintf("https://ygoprodeck.com/api/pack/setSearch.php?cardset=%s&region=MD", packName))
	if err != nil {
		log.Fatalln(err)
	}

	secretPackCardsBytes, err := ioutil.ReadAll(secretPackCardsResponse.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var secretPackCards []SecretPackCard
	err = json.Unmarshal(secretPackCardsBytes, &secretPackCards)
	if err != nil {
		log.Fatalln(err)
	}

	secretPackCards = fixPack(packName, secretPackCards)

	secretPackCardMap := make(map[Rarity][]SecretPackCard)
	for _, card := range secretPackCards {
		secretPackCardMap[card.CardVariations[0].CardRarity] = append(secretPackCardMap[card.CardVariations[0].CardRarity], card)
	}

	alsoPullFromMasterPack := true
	if len(secretPackCardMap[RarityUltraRare]) >= 13 {
		alsoPullFromMasterPack = false
	}

	log.Println("Finished fetching secret pack cards")

	// Get the master pack cards
	masterPackCardsResponse, err := http.Get("https://db.ygoprodeck.com/queries/master_duel/getMasterDuel.php")
	if err != nil {
		log.Fatalln(err)
	}

	masterPackCardsBytes, err := ioutil.ReadAll(masterPackCardsResponse.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var masterPackCards MasterPackQuery
	err = json.Unmarshal(masterPackCardsBytes, &masterPackCards)
	if err != nil {
		log.Fatalln(err)
	}

	masterPackCardMap := make(map[Rarity][]MasterPackCard)
	for _, card := range masterPackCards.Data {
		masterPackCardMap[card.Rarity] = append(masterPackCardMap[card.Rarity], card)
	}

	log.Println("Finished fetching master pack cards")

	pulls := getPulls()

	result := make([][]ResultCard, len(pulls))
	numURs := 0
	for i := 0; i < len(pulls); i++ {
		log.Println("=================")
		log.Printf("     Pack %d    ", i+1)
		log.Println("=================")

		result[i] = make([]ResultCard, len(pulls[i]))

		for j := 0; j < len(pulls[i]); j++ {
			pulledCard := pulls[i][j]
			if pulledCard.Rarity == RarityUltraRare {
				numURs++
			}

			if j < 4 && alsoPullFromMasterPack {
				// Pull from the master pack
				cardIndex := rand.Intn(len(masterPackCardMap[pulledCard.Rarity]))
				masterPackCard := masterPackCardMap[pulledCard.Rarity][cardIndex]
				result[i][j] = ResultCard{
					CardName:   masterPackCard.Name,
					CardID:     0,  // TODO
					CardImg:    "", // TODO
					CardRarity: pulledCard.Rarity,
					CardFoil:   pulledCard.Foil,
				}
			} else {
				// Pull from the secret pack
				cardIndex := rand.Intn(len(secretPackCardMap[pulledCard.Rarity]))
				secretPackCard := secretPackCardMap[pulledCard.Rarity][cardIndex]
				result[i][j] = ResultCard{
					CardName:   secretPackCard.CardName,
					CardID:     secretPackCard.CardID,
					CardImg:    secretPackCard.CardImg,
					CardRarity: pulledCard.Rarity,
					CardFoil:   pulledCard.Foil,
				}
			}

			logText := fmt.Sprintf("%d. ", j+1)
			if pulledCard.Foil == FoilGlossy {
				logText += "~~glossy~~ "
			} else if pulledCard.Foil == FoilRoyal {
				logText += "$$$$ ROYAL $$$$ "
			}
			logText += fmt.Sprintf("%s (%s)", result[i][j].CardName, result[i][j].CardRarity)
			if pulledCard.Rarity == RaritySuperRare {
				logText = GreenString(logText)
			} else if pulledCard.Rarity == RarityUltraRare {
				logText = RedString(logText)
			}
			log.Println(logText)
		}
	}

	log.Println("**************")
	log.Printf("Got %d URs", numURs)
}

func getPulls() [][]PulledCard {
	rChanceA := 35.0
	srChanceA := 7.5
	urChanceA := 2.5

	srChanceB := 7.5
	urChanceB := 2.5

	urChanceC := 20.0

	pulls := make([][]PulledCard, 10)

	for i := 0; i < 10; i++ {
		pulls[i] = make([]PulledCard, 8)
		for j := 0; j < 8; j++ {
			var card PulledCard
			roll := rand.Float64() * 100
			if j < 7 {
				if roll < urChanceA {
					card = getFoil(RarityUltraRare)
				} else if roll < urChanceA+srChanceA {
					card = getFoil(RaritySuperRare)
				} else if roll < urChanceA+srChanceA+rChanceA {
					card = getFoil(RarityRare)
				} else {
					card = getFoil(RarityCommon)
				}
			} else {
				if i < 9 {
					if roll < urChanceB {
						card = getFoil(RarityUltraRare)
					} else if roll < urChanceB+srChanceB {
						card = getFoil(RaritySuperRare)
					} else {
						card = getFoil(RarityRare)
					}
				} else {
					if roll < urChanceC {
						card = getFoil(RarityUltraRare)
					} else {
						card = getFoil(RaritySuperRare)
					}
				}
			}

			pulls[i][j] = card
		}
	}

	return pulls
}

func getFoil(rarity Rarity) PulledCard {
	foilRoyalChance := 1.0
	foilGlossyChance := 9.0

	roll := rand.Float64() * 100

	var foil Foil
	if roll < foilRoyalChance && (rarity == RarityUltraRare || rarity == RaritySuperRare) {
		foil = FoilRoyal
	} else if roll < foilRoyalChance+foilGlossyChance {
		foil = FoilGlossy
	} else {
		foil = FoilNone
	}

	return PulledCard{
		Rarity: rarity,
		Foil:   foil,
	}
}

const colorRed = "\033[0;31m"
const colorGreen = "\033[0;32m"

// const colorBlue = "\033[0;34m"
const colorNone = "\033[0m"

// RedString returns a red string.
func RedString(s string) string {
	return colorRed + s + colorNone
}

// GreenString returns a green string.
func GreenString(s string) string {
	return colorGreen + s + colorNone
}

func addCardIfNotExists(cardName string, rarity Rarity, cards []SecretPackCard) []SecretPackCard {
	for _, card := range cards {
		if card.CardName == cardName {
			return cards
		}
	}
	cards = append(cards, SecretPackCard{
		CardName: cardName,
		CardVariations: []SecretPackCardVariation{
			{
				CardRarity: rarity,
			},
		},
	})
	return cards
}

func fixPack(packName string, secretPackCards []SecretPackCard) []SecretPackCard {
	lenBefore := len(secretPackCards)
	if packName == "Singular Strike Overthrow" {
		secretPackCards = addCardIfNotExists("Surgical Striker - H.A.M.P.", RarityUltraRare, secretPackCards)
		secretPackCards = addCardIfNotExists("Mathmech Circular", RarityUltraRare, secretPackCards)
		secretPackCards = addCardIfNotExists("Sky Striker Mobilize - Linkage!", RarityUltraRare, secretPackCards)
		secretPackCards = addCardIfNotExists("Aileron", RaritySuperRare, secretPackCards)
	}
	lenAfter := len(secretPackCards)
	if lenBefore != lenAfter {
		log.Printf("Added %d cards to secret pack", lenAfter-lenBefore)
	} else {
		log.Println("No fixes to secret pack made")
	}
	return secretPackCards
}
