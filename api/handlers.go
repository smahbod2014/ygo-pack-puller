package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type PerformPullsRequest struct {
	PackName string `json:"pack_name"`
	NumPacks int    `json:"num_packs"`
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
	secretPackCardsResponse, err := http.Get(fmt.Sprintf("https://ygoprodeck.com/api/pack/setSearch.php?cardset=%s&region=MD", url.QueryEscape(requestBody.PackName)))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	secretPackCardsBytes, err := io.ReadAll(secretPackCardsResponse.Body)
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
	if isSelectionPack(requestBody.PackName) || requestBody.PackName == "Master Pack" {
		alsoPullFromMasterPack = false
	}

	var masterPackCards []SecretPackCard
	if alsoPullFromMasterPack {
		// Get the master pack cards
		masterPackCardsResponse, err := http.Get(fmt.Sprintf("https://ygoprodeck.com/api/pack/setSearch.php?cardset=%s&region=MD", url.QueryEscape("Master Pack")))
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		masterPackCardsBytes, err := io.ReadAll(masterPackCardsResponse.Body)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		err = json.Unmarshal(masterPackCardsBytes, &masterPackCards)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	masterPackCardMap := make(map[Rarity][]SecretPackCard)
	for _, card := range masterPackCards {
		masterPackCardMap[card.CardVariations[0].CardRarity] = append(masterPackCardMap[card.CardVariations[0].CardRarity], card)
	}

	pulls := getPullRarities(requestBody.NumPacks)

	result := make([][]ResultCard, len(pulls))
	numURs := 0
	for i := 0; i < len(pulls); i++ {
		result[i] = make([]ResultCard, len(pulls[i]))
		for j := 0; j < len(pulls[i]); j++ {
			pulledCard := pulls[i][j]
			if pulledCard.Rarity == RarityUltraRare {
				numURs++
			}

			var selectedCard SecretPackCard
			if j < 4 && alsoPullFromMasterPack {
				// Pull from the master pack
				cardIndex := rand.Intn(len(masterPackCardMap[pulledCard.Rarity]))
				selectedCard = masterPackCardMap[pulledCard.Rarity][cardIndex]
			} else {
				// Pull from the secret pack
				cardIndex := rand.Intn(len(secretPackCardMap[pulledCard.Rarity]))
				selectedCard = secretPackCardMap[pulledCard.Rarity][cardIndex]
			}

			result[i][j] = ResultCard{
				CardName:   selectedCard.CardName,
				CardID:     selectedCard.CardID,
				CardImg:    selectedCard.CardImg,
				CardRarity: pulledCard.Rarity,
				CardFoil:   pulledCard.Foil,
			}
		}
	}

	ctx.JSON(http.StatusOK, PerformPullsResponse{
		PackName: requestBody.PackName,
		NumURs:   numURs,
		Pulls:    result,
	})
}

func getPullRarities(numPacks int) [][]PulledCard {
	rChanceA := 35.0
	srChanceA := 7.5
	urChanceA := 2.5

	srChanceB := 7.5
	urChanceB := 2.5

	urChanceC := 20.0

	pulls := make([][]PulledCard, numPacks)

	for i := 0; i < numPacks; i++ {
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
				if i%10 < 9 {
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

func removeCardIfExists(cardName string, cards []SecretPackCard) []SecretPackCard {
	var fixedCards []SecretPackCard
	for _, card := range cards {
		if card.CardName != cardName {
			fixedCards = append(fixedCards, card)
		}
	}
	return fixedCards
}

func fixPack(packName string, secretPackCards []SecretPackCard) []SecretPackCard {
	if packName == "Singular Strike Overthrow" {
		secretPackCards = addCardIfNotExists("Surgical Striker - H.A.M.P.", RarityUltraRare, secretPackCards)
		secretPackCards = addCardIfNotExists("Mathmech Circular", RarityUltraRare, secretPackCards)
		secretPackCards = addCardIfNotExists("Sky Striker Mobilize - Linkage!", RarityUltraRare, secretPackCards)
		secretPackCards = addCardIfNotExists("Aileron", RaritySuperRare, secretPackCards)
	} else if packName == "Blazing Fortitude" {
		secretPackCards = removeCardIfExists("Stall Turn", secretPackCards)
	}

	if packName != "Rulers of the Deep" && packName != "Master Pack" {
		secretPackCards = removeCardIfExists("Fury of Kairyu-Shin", secretPackCards)
	}
	return secretPackCards
}

func isSelectionPack(packName string) bool {
	return packName == "Revival of Legends" ||
		packName == "Stalwart Force" ||
		packName == "Ruler's Mask" ||
		packName == "Fusion Potential" ||
		packName == "Refined Blade" ||
		packName == "Valiant Wings" ||
		packName == "Wandering Travelers" ||
		packName == "Invincible Raid" ||
		packName == "The Newborn Dragon" ||
		packName == "Cosmic Ocean" ||
		packName == "Battle Trajectory" ||
		packName == "Mysterious Labyrinth" ||
		packName == "Beginning of Turmoil" ||
		packName == "Heroic Warriors" ||
		packName == "Recollection of Stories" ||
		packName == "Beyond Speed"
}
