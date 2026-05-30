import type { Route } from "./+types/layout";
import { Welcome } from "../welcome/welcome";
import { Link, Outlet } from "react-router";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "New React Router App" },
    { name: "description", content: "Welcome to React Router!" },
  ];
}

export default function Layout() {
  return <div>
    <Welcome />
    <div className="flex items-center justify-center">
      <Link to="sub1">Sub1</Link>
      &nbsp; | &nbsp;
      <Link to="sub2">Sub2</Link>
    </div>
    <Outlet />
  </div>;
}
