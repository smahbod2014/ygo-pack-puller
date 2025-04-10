package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/carlmjohnson/versioninfo"
	"github.com/gin-gonic/gin"
	"github.com/jucardi/go-streams/v2/streams"
	"github.com/pkg/errors"
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

func PerformPullsHandler(ctx *gin.Context) {
	requestBody := PerformPullsRequest{}
	err := ctx.BindJSON(&requestBody)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	packs, err := getPacks()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	pack := streams.
		From[MDMPack](packs).
		Filter(func(pack MDMPack) bool {
			return pack.Name == requestBody.PackName
		}).
		First()

	if pack.ID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Pack not found",
		})
		return
	}

	cards, err := fetchAllCardsFromPack(pack.ID)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	seenCards := map[string]bool{}
	cardMapByRarity := map[Rarity][]MDMCard{}
	for _, card := range cards {
		if seenCards[card.Name] {
			continue
		}
		seenCards[card.Name] = true
		cardMapByRarity[card.Rarity] = append(cardMapByRarity[card.Rarity], card)
	}

	alsoPullFromMasterPack := pack.Type == MDMPackTypeSecretPack

	masterPackCardMapByRarity := map[Rarity][]MDMCard{}
	if alsoPullFromMasterPack {
		masterPack := streams.
			From[MDMPack](packs).
			Filter(func(pack MDMPack) bool {
				return pack.Type == MDMPackTypeMasterPack
			}).
			First()

		if masterPack.ID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "Master Pack not found",
			})
			return
		}
		masterPackCards, err := fetchAllCardsFromPack(masterPack.ID)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		for k := range seenCards {
			delete(seenCards, k)
		}
		for _, card := range masterPackCards {
			if seenCards[card.Name] {
				continue
			}
			seenCards[card.Name] = true
			masterPackCardMapByRarity[card.Rarity] = append(masterPackCardMapByRarity[card.Rarity], card)
		}
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

			var selectedCard MDMCard
			if j < 4 && alsoPullFromMasterPack {
				// Pull from the master pack
				cardIndex := rand.Intn(len(masterPackCardMapByRarity[pulledCard.Rarity]))
				selectedCard = masterPackCardMapByRarity[pulledCard.Rarity][cardIndex]
			} else {
				// Pull from the chosen pack
				cardIndex := rand.Intn(len(cardMapByRarity[pulledCard.Rarity]))
				selectedCard = cardMapByRarity[pulledCard.Rarity][cardIndex]
			}

			result[i][j] = ResultCard{
				CardName:   selectedCard.Name,
				CardID:     strconv.Itoa(int(selectedCard.KonamiID)),
				CardImg:    "https://s3.duellinksmeta.com/cards/" + selectedCard.ID + "_w420.webp",
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

func fetchAllCardsFromPack(packID string) ([]MDMCard, error) {
	var cards []MDMCard
	apiURL := "https://www.masterduelmeta.com/api/v1/cards?obtain.source=%s&cardSort=monsterTypeOrder&aggregate=search&fields=name,rarity,konamiID&page=%d&limit=1000"
	page := 1
	for {
		fullURL := fmt.Sprintf(apiURL, url.QueryEscape(packID), page)
		log.Printf("Making API call: GET to %s", fullURL)
		response, err := http.Get(fullURL)
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch cards from cards API")
		}

		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read response body from cards API")
		}

		var readCards []MDMCard
		err = json.Unmarshal(bytes, &readCards)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal cards from response body: %s", string(bytes)))
		}

		if len(readCards) == 0 {
			break
		}

		var fixedReadCards []MDMCard
		for _, card := range readCards {
			if card.KonamiID == 0 {
				if card.Name == "GranSolfacord Coolia" {
					// masterduelmeta has a typo in the name...
					card.Name = "GranSolfachord Coolia"
				}

				// Get the konami ID from ygoprodeck since masterduelmeta doesn't have it
				ygoprodeckResponse, err := http.Get("https://db.ygoprodeck.com/api/v7/cardinfo.php?name=" + url.QueryEscape(card.Name))
				if err != nil {
					return nil, errors.Wrap(err, "failed to get fixed card from ygoprodeck")
				}

				// get ID from the first array element in the json response body
				ygoprodeckBytes, err := io.ReadAll(ygoprodeckResponse.Body)
				if err != nil {
					return nil, errors.Wrap(err, "failed to read response body from ygoprodeck")
				}

				var unmarshalledResponse YGOProDeckResponse
				err = json.Unmarshal(ygoprodeckBytes, &unmarshalledResponse)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshal ygoprodeck response for %s: %s", card.Name, string(ygoprodeckBytes)))
				}

				card.KonamiID = FlexInt(unmarshalledResponse.Data[0].ID)
			}
			fixedReadCards = append(fixedReadCards, card)
		}

		cards = append(cards, fixedReadCards...)
		page++
	}

	return cards, nil
}

func getPacks() ([]MDMPack, error) {
	setsResponse, err := http.Get("https://www.masterduelmeta.com/api/v1/sets?page=1&limit=500&fields=name,release,type")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get master duel sets")
	}

	bytes, err := io.ReadAll(setsResponse.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read master duel sets response body")
	}

	var allPacks []MDMPack
	err = json.Unmarshal(bytes, &allPacks)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal master duel sets bytes")
	}

	return streams.
		From[MDMPack](allPacks).
		Filter(func(pack MDMPack) bool {
			return pack.Type == MDMPackTypeMasterPack ||
				pack.Type == MDMPackTypeSecretPack ||
				pack.Type == MDMPackTypeSelectionPack
		}).
		ToArray(), nil
}

func GetPacksHandler(ctx *gin.Context) {
	packs, err := getPacks()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, packs)
}

func GetGitCommitHashHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"hash": versioninfo.Revision,
		"date": versioninfo.LastCommit,
	})
}
