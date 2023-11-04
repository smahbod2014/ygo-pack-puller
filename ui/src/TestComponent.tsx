import { Button, Form, Modal } from "react-bootstrap";
import "./TestComponent.css";
import { useState } from "react";

export const TestComponent = () => {
  return (
    <div className="TestComponentOuter">
      <div className="TestComponentInner">
        <div className="spacer" />
        <h1>Random Draft Mode</h1>
        <div className="spacer" />

        <Form className="RoomContainer">
          <Form.Group controlId="formRoomID">
            <Form.Control size="lg" type="text" autoComplete="off" className="RoomTextInput" />
          </Form.Group>
          <Button className="RoomButton" variant="primary">
            Create/Join Room
          </Button>
        </Form>
      </div>
    </div>
  );
};
