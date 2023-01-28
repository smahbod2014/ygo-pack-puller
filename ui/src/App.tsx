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

function App() {
  const packOptions = secretPacks.sort().map((e) => ({ label: e, value: e }));
  const [selectedPack, setSelectedPack] = useState<string>(packOptions[0].value);
  const [performPullsResponse, setPerformPullsResponse] = useState<PerformPullsResponse | undefined>(undefined);
  const [currentPackNumber, setCurrentPackNumber] = useState(0);
  const [revealedCards, setRevealedCards] = useState(new Array<boolean>(10));
  const [numPacksToPull, setNumPacksToPull] = useState(10);
  const [nextPackLoading, setNextPackLoading] = useState(false);

  const postPullPacks = async () => {
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
    setRevealedCards(new Array<boolean>(10));
  };

  const updateRevealedCards = (i: number, flipped: boolean) => {
    const copy = [...revealedCards];
    copy[i] = flipped;
    setRevealedCards(copy);
  };

  return (
    <div className="App">
      <header className="App-header">
        <div className="PackSelectorContainer">
          <div className="PackSelectorRow">
            <p>Choose a pack</p>
            <Select className="PackSelector" options={packOptions} defaultValue={packOptions[0]} onChange={(e) => setSelectedPack(e!.value)} />
          </div>
          <div className="PackSelectorRow">
            <p>Number of packs</p>
            <NumberPicker className="PackSelectorNumberPicker" defaultValue={10} min={1} max={30} onChange={(value) => setNumPacksToPull(value ?? 10)} />
          </div>
          <div className="PackSelectorRow">
            <Button className="PackPullButton" variant="warning" onClick={() => postPullPacks()}>
              Pull packs
            </Button>
          </div>
        </div>

        {performPullsResponse && performPullsResponse.pulls.length > 0 && (
          <>
            <p>Pack {currentPackNumber + 1} of 10</p>
            <div className={classNames("PullsContainer", nextPackLoading && "Hidden")}>
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
                      <img className="CardImage" src={card.card_img} draggable={false} />
                    </div>
                  </div>
                </div>
              ))}
            </div>
            {nextPackLoading && <p className="Loading">Loading...</p>}
            <Button
              className={classNames("PackPullButton", currentPackNumber === 9 && "Hidden")}
              style={{ marginRight: "7.5%" }}
              variant="success"
              onClick={() => {
                setNextPackLoading(true);
                setTimeout(() => {
                  setNextPackLoading(false);
                }, 500);
                if (currentPackNumber < 9) {
                  setCurrentPackNumber(currentPackNumber + 1);
                  setRevealedCards(new Array<boolean>(10));
                }
              }}
            >
              Next pack
            </Button>
          </>
        )}
      </header>
    </div>
  );
}

export default App;
