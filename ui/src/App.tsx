import React, { useCallback, useState } from "react";
import logo from "./logo.svg";
import "./App.css";
import "react-widgets/styles.css";
import Select from "react-select";
import { Button } from "react-bootstrap";
import { json } from "stream/consumers";
import { cardBackImage, rarityImages, rarityURImage, secretPacks } from "./Data";
import classNames from "classnames";
import NumberPicker from "react-widgets/NumberPicker";
import Tooltip from "@mui/material/Tooltip";
import Sparkle from "react-sparkle";

interface PerformPullsResponse {
  pack_name: string;
  num_urs: number;
  pulls: ResultCard[][];
}

interface ResultCard {
  card_id: number;
  card_name: string;
  card_img: string;
  card_rarity: string;
  card_foil: string;
}

function getCardRarityAsNumber(card: ResultCard) {
  switch (card.card_rarity) {
    case "Common":
      return 0;
    case "Rare":
      return 1;
    case "Super Rare":
      return 2;
    case "Ultra Rare":
      return 3;
    default:
      return 4;
  }
}

function compareByRarity(a: ResultCard, b: ResultCard): number {
  return getCardRarityAsNumber(a) - getCardRarityAsNumber(b);
}

function App() {
  const packOptions = secretPacks.sort().map((e) => ({ label: e, value: e }));
  const [selectedPack, setSelectedPack] = useState<string>(packOptions[0].value);
  const [performPullsResponse, setPerformPullsResponse] = useState<PerformPullsResponse | undefined>(undefined);
  const [currentPackNumber, setCurrentPackNumber] = useState(0);
  const [revealedCards, setRevealedCards] = useState(new Array<boolean>(8));
  const [numPacksToPull, setNumPacksToPull] = useState(10);
  const [numPacksInCurrentPull, setNumPacksInCurrentPull] = useState(10);
  const [nextPackLoading, setNextPackLoading] = useState(false);
  const [showingSummary, setShowingSummary] = useState(false);

  const postPullPacks = async () => {
    setNextPackLoading(true);

    const jsonResponse: PerformPullsResponse = await fetch("/api/pull", {
      method: "POST",
      body: JSON.stringify({
        pack_name: selectedPack,
        num_packs: numPacksToPull,
      }),
      headers: { "Content-Type": "application/json" },
    }).then((result) => result.json());

    setPerformPullsResponse(jsonResponse);
    setCurrentPackNumber(0);
    setRevealedCards(new Array<boolean>(8));
    setNextPackLoading(false);
    setNumPacksInCurrentPull(numPacksToPull);
  };

  const updateRevealedCards = (i: number, flipped: boolean) => {
    const copy = [...revealedCards];
    copy[i] = flipped;
    setRevealedCards(copy);
  };

  if (showingSummary && performPullsResponse) {
    const sortedCards = performPullsResponse.pulls.flat().sort((a, b) => {
      const rarityComparison = -compareByRarity(a, b);
      if (rarityComparison !== 0) {
        return rarityComparison;
      }
      return a.card_name.localeCompare(b.card_name);
    });

    const downloadBanlistFile = () => {
      const cardCounts = new Map<number, number>();
      sortedCards.forEach((card) => {
        cardCounts.set(card.card_id, Math.min((cardCounts.get(card.card_id) ?? 0) + 1, 3));
      });
      let fileContents = "!Custom Banlist\n";
      fileContents += "$whitelist\n";
      cardCounts.forEach((count, id) => {
        fileContents += id.toString() + " " + count.toString() + "\n";
      });
      const element = document.createElement("a");
      const file = new Blob([fileContents], { type: "text/plain;charset=utf-8" });
      element.href = URL.createObjectURL(file);
      element.download = "Banlist " + new Date().toISOString().replace("T", " ").split(".")[0] + "Z.conf";
      document.body.appendChild(element);
      element.click();
    };
    return (
      <div className="App">
        <header className="App-header">
          <div className="SummaryButtonContainer">
            <Button className="SummaryButton" variant="success" onClick={downloadBanlistFile}>
              Download as banlist
            </Button>
            <Button
              className="SummaryButton"
              variant="success"
              onClick={() => {
                setShowingSummary(false);
                setPerformPullsResponse(undefined);
              }}
            >
              Done
            </Button>
          </div>

          <div className="SummaryContainer">
            {[...Array(sortedCards.length / 8)].map((cardRow, i) => {
              return (
                <div className="SummaryRow" key={i}>
                  {sortedCards.slice(i * 8, i * 8 + 8).map((card, j) => {
                    return (
                      <div className="SummaryCard">
                        <img className="CardRarity" src={rarityImages[card.card_rarity]} draggable={false} />
                        <a href={"https://ygoprodeck.com/card?search=" + card.card_id} target="_blank" draggable={false}>
                          <div className={card.card_foil} style={{ position: "relative" }}>
                            {card.card_foil === "royal" && <Sparkle minSize={20} maxSize={30} color={"random"} fadeOutSpeed={30} />}
                            <img className="CardImage" src={card.card_img} draggable={false} />
                          </div>
                        </a>
                      </div>
                    );
                  })}
                </div>
              );
            })}
          </div>
        </header>
      </div>
    );
  }

  return (
    <div className="App">
      <header className="App-header">
        <div className="PackSelectorContainer">
          <div className="PackSelectorRow">
            <p>Choose a pack</p>
            <Select
              className="PackSelector"
              options={packOptions}
              defaultValue={{ label: selectedPack, value: selectedPack }}
              onChange={(e) => setSelectedPack(e!.value)}
            />
          </div>
          <div className="PackSelectorRow">
            <p>Number of packs</p>
            <NumberPicker
              className="PackSelectorNumberPicker"
              defaultValue={numPacksInCurrentPull}
              min={1}
              max={30}
              onChange={(value) => setNumPacksToPull(value ?? 10)}
            />
          </div>
          <div className="PackSelectorRow">
            <p style={{ color: "gray", fontSize: "0.5em" }}>Guaranteed SR or higher every 10th pack!</p>
            <Button className="StartPullingPacksButton" variant="warning" onClick={() => postPullPacks()}>
              Pull packs
            </Button>
          </div>
        </div>

        {performPullsResponse && performPullsResponse.pulls.length > 0 && (
          <>
            <p>
              Pack {currentPackNumber + 1} of {numPacksInCurrentPull}
            </p>
            {!nextPackLoading ? (
              <div className="PullsContainer">
                {performPullsResponse.pulls[currentPackNumber].map((card, i) => (
                  <div
                    className={classNames("flip-card", revealedCards[i] && "flipped")}
                    key={i}
                    onMouseOver={(e) => e.buttons === 1 && updateRevealedCards(i, true)}
                    onMouseDown={() => updateRevealedCards(i, true)}
                  >
                    <div className="flip-card-inner">
                      <div className="flip-card-front">
                        <img className="CardRarity Hidden" src={rarityImages[card.card_rarity]} draggable={false} />
                        <img className="CardImage" src={cardBackImage} draggable={false} />
                      </div>
                      <div className="flip-card-back">
                        <img className="CardRarity" src={rarityImages[card.card_rarity]} draggable={false} />
                        <a href={"https://ygoprodeck.com/card?search=" + card.card_id} target="_blank" draggable={false}>
                          <div className={card.card_foil} style={{ position: "relative" }}>
                            {card.card_foil === "royal" && <Sparkle minSize={20} maxSize={30} color={"random"} fadeOutSpeed={30} />}
                            <img className="CardImage" src={card.card_img} draggable={false} />
                          </div>
                        </a>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="Loading">Loading...</p>
            )}

            <div className="PackPullButtonContainer">
              {currentPackNumber < numPacksInCurrentPull - 1 ? (
                <>
                  <Button
                    className="PackPullButton"
                    variant="success"
                    onClick={() => {
                      setShowingSummary(true);
                    }}
                  >
                    Skip to summary
                  </Button>
                  <Button
                    className="PackPullButton"
                    variant="success"
                    onClick={() => {
                      setNextPackLoading(true);
                      setTimeout(() => {
                        setNextPackLoading(false);
                        setTimeout(() => {
                          setRevealedCards(Array(8).fill(true));
                        }, 500);
                      }, 500);
                      setCurrentPackNumber(currentPackNumber + 1);
                      setRevealedCards(new Array<boolean>(8));
                    }}
                  >
                    Auto open next pack
                  </Button>
                  <Button
                    className="PackPullButton"
                    variant="success"
                    onClick={() => {
                      setNextPackLoading(true);
                      setTimeout(() => {
                        setNextPackLoading(false);
                      }, 500);
                      setCurrentPackNumber(currentPackNumber + 1);
                      setRevealedCards(new Array<boolean>(8));
                    }}
                  >
                    Next pack
                  </Button>
                </>
              ) : (
                <Button
                  className="PackPullButton"
                  variant="success"
                  onClick={() => {
                    setShowingSummary(true);
                  }}
                >
                  Summary
                </Button>
              )}
            </div>
          </>
        )}
      </header>
    </div>
  );
}

export default App;
