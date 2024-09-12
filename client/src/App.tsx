import { Outlet } from "react-router-dom";
import "./App.css";
import { Button } from "@mantine/core";
import Login from "./features/Auth/Login/Login";

function App() {
  return (
    <>
      {" "}
      <div id="sidebar">
        <Login />

        <div>
          <form id="search-form" role="search">
            <input
              id="q"
              aria-label="Search contacts"
              placeholder="Search"
              type="search"
              name="q"
            />
            <div id="search-spinner" aria-hidden hidden={true} />
            <div className="sr-only" aria-live="polite"></div>
          </form>
          <form method="post">
            <Button type="submit">New</Button>
          </form>
        </div>
        <nav>
          <ul>
            <li>
              <a href={`/contacts/1`}>Your Name</a>
            </li>
            <li>
              <a href={`/contacts/2`}>Your Friend</a>
            </li>
          </ul>
        </nav>
      </div>
      <div id="detail">
        <Outlet />
      </div>
    </>
  );
}

export default App;
