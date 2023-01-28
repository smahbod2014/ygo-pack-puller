package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thoas/go-funk"
	"golang.org/x/exp/maps"
)

type PerformPullsRequest struct {
	PackName string `json:"pack_name"`
}

type PerformPullsResponse struct {
	PackName string         `json:"pack_name"`
	NumURs   int            `json:"num_urs"`
	Pulls    [][]ResultCard `json:"pulls"`
}

func performPulls(ctx *gin.Context) {
	requestBody := PerformPullsRequest{}
	err := ctx.BindJSON(&requestBody)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Get the cards from the pack
	secretPackCardsResponse, err := http.Get(fmt.Sprintf("https://ygoprodeck.com/api/pack/setSearch.php?cardset=%s&region=MD", requestBody.PackName))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	secretPackCardsBytes, err := ioutil.ReadAll(secretPackCardsResponse.Body)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var secretPackCards []SecretPackCard
	err = json.Unmarshal(secretPackCardsBytes, &secretPackCards)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	secretPackCards = fixPack(requestBody.PackName, secretPackCards)

	secretPackCardMap := make(map[Rarity][]SecretPackCard)
	for _, card := range secretPackCards {
		secretPackCardMap[card.CardVariations[0].CardRarity] = append(secretPackCardMap[card.CardVariations[0].CardRarity], card)
	}

	alsoPullFromMasterPack := true
	if len(secretPackCardMap[RarityUltraRare]) >= 13 {
		alsoPullFromMasterPack = false
	}

	// Get the master pack cards
	masterPackCardsResponse, err := http.Get("https://db.ygoprodeck.com/queries/master_duel/getMasterDuel.php")
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	masterPackCardsBytes, err := ioutil.ReadAll(masterPackCardsResponse.Body)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var masterPackCards MasterPackQuery
	err = json.Unmarshal(masterPackCardsBytes, &masterPackCards)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	masterPackCardMap := make(map[Rarity][]MasterPackCard)
	for _, card := range masterPackCards.Data {
		if card.Pack == "Legacy Pack" {
			continue
		}
		masterPackCardMap[card.Rarity] = append(masterPackCardMap[card.Rarity], card)
	}

	pulls := getPullRarities()

	masterPackCardNames := map[string]bool{}
	result := make([][]ResultCard, len(pulls))
	numURs := 0
	for i := 0; i < len(pulls); i++ {
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
				masterPackCardNames[masterPackCard.Name] = true
				result[i][j] = ResultCard{
					CardName:   masterPackCard.Name,
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
		}
	}

	cardInfo, err := GetCardInfo(maps.Keys(masterPackCardNames))
	if err != nil {
		log.Println("Failed to get card info", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(cardInfo.Data) != len(masterPackCardNames) {
		log.Println("Mismatched card info. Master pack card names", masterPackCardNames)
		log.Println("Queried card data", cardInfo.Data)
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("CardInfoData length mismatch, expected %d, got %d", len(masterPackCardNames), len(cardInfo.Data)))
		return
	}

	cardInfoByName := funk.Map(cardInfo.Data, func(cardInfoData CardInfoData) (string, CardInfoData) {
		return cardInfoData.Name, cardInfoData
	}).(map[string]CardInfoData)

	for i, packCards := range result {
		for j, card := range packCards {
			if card.CardID == 0 {
				result[i][j].CardID = cardInfoByName[card.CardName].ID
				result[i][j].CardImg = cardInfoByName[card.CardName].CardImages[0].ImageURL
			}
		}
	}

	ctx.JSON(http.StatusOK, PerformPullsResponse{
		PackName: requestBody.PackName,
		NumURs:   numURs,
		Pulls:    result,
	})
}

func getPullRarities() [][]PulledCard {
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
	if packName == "Singular Strike Overthrow" {
		secretPackCards = addCardIfNotExists("Surgical Striker - H.A.M.P.", RarityUltraRare, secretPackCards)
		secretPackCards = addCardIfNotExists("Mathmech Circular", RarityUltraRare, secretPackCards)
		secretPackCards = addCardIfNotExists("Sky Striker Mobilize - Linkage!", RarityUltraRare, secretPackCards)
		secretPackCards = addCardIfNotExists("Aileron", RaritySuperRare, secretPackCards)
	}
	return secretPackCards
}
