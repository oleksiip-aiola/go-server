import { Outlet } from "react-router-dom";
import "./App.css";
import Login from "./features/Auth/Login/Login";

function App() {
  return (
    <>
      <Login />
      <Outlet />
    </>
  );
}

export default App;
