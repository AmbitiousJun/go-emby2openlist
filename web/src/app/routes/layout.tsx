import { House, Menu, Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";
import { Outlet, useLocation, useNavigate } from "react-router";
import SettingsModal from "~/components/settings_modal/settings_modal";
import type { Route } from "./+types/layout";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Ge2o Web" },
    { name: "description", content: "go-emby2openlist web application" },
  ];
}

export type ThemeContext = {
  dark: boolean;
  setDark: (dark: boolean) => void;
};

type NavItem = {
  label: string;
  to?: string;
  children: NavItem[];
};

const navData: NavItem[] = [
  {
    label: "接口调用",
    children: [
      {
        label: "OpenList 本地目录树",
        to: "/api/openlist_local_tree",
        children: [],
      },
    ],
  },
  {
    label: "日志",
    to: "/log",
    children: [],
  },
];

export default function Layout() {
  const navigate = useNavigate();
  const location = useLocation();
  const [dark, setDark] = useState(localStorage.getItem("dark") ? true : false);
  const [menuOpen, setMenuOpen] = useState(false);

  // 自动切换主题色
  useEffect(() => {
    document.documentElement.setAttribute(
      "data-theme",
      dark ? "dark" : "light",
    );
  }, [dark]);

  // 监听页面路径变化 自动收起导航栏
  useEffect(() => {
    document
      .querySelectorAll("details")
      .forEach((item) => ((item as HTMLDetailsElement).open = false));
  }, [location.pathname]);

  const setDarkAndSave = (dark: boolean) => {
    setDark(dark);
    localStorage.setItem("dark", dark ? "1" : "");
  };

  const navigateToMediaServerHome = () => {
    window.location.href = `${window.location.origin}/`;
  };

  const navItemMapperHorizotal = (item: NavItem) => {
    if (item.children.length <= 0) {
      return (
        <li>
          <button onClick={() => navigate(item.to ?? "/")}>{item.label}</button>
        </li>
      );
    }
    return (
      <li>
        <details>
          <summary>{item.label}</summary>
          <ul className="p-2 bg-base-100 w-60 z-1">
            {item.children.map((itemInner) =>
              navItemMapperHorizotal(itemInner),
            )}
          </ul>
        </details>
      </li>
    );
  };

  const navItemMapperVertical = (item: NavItem) => {
    if (item.children.length <= 0) {
      return (
        <li>
          <button
            onClick={() => {
              setMenuOpen(false);
              navigate(item.to ?? "/");
            }}
          >
            {item.label}
          </button>
        </li>
      );
    }

    return (
      <li>
        <button>{item.label}</button>
        <ul>
          {item.children.map((itemInner) => navItemMapperVertical(itemInner))}
        </ul>
      </li>
    );
  };

  return (
    <div>
      <nav>
        <div className="max-lg:collapse bg-base-200 mb-12 shadow-sm w-full rounded-md px-4">
          {/* 全屏遮罩 */}
          <input
            id="navbar-1-toggle"
            className="peer hidden"
            type="checkbox"
            checked={menuOpen}
            onChange={(e) => setMenuOpen(e.target.checked)}
          />
          <label
            htmlFor="navbar-1-toggle"
            className="fixed inset-0 hidden max-lg:peer-checked:block"
          ></label>

          {/* 导航栏主体 */}
          <div className="collapse-title navbar">
            {/* 导航栏左侧 */}
            <div className="navbar-start">
              <label
                htmlFor="navbar-1-toggle"
                className="btn btn-ghost lg:hidden"
              >
                <Menu />
              </label>
              <button
                className="btn btn-ghost text-xl"
                onClick={() => navigate("/")}
              >
                Ge2o Web
              </button>
            </div>

            {/* 导航栏内容 */}
            <div className="navbar-center hidden lg:flex">
              <ul className="menu menu-horizontal px-1">
                {navData.map((item) => navItemMapperHorizotal(item))}
              </ul>
            </div>

            {/* 导航栏右侧 */}
            <div className="navbar-end space-x-4">
              {/* 主题切换 */}
              <label className="swap swap-rotate">
                {/* this hidden checkbox controls the state */}
                <input
                  type="checkbox"
                  checked={dark}
                  onChange={(e) => setDarkAndSave(e.target.checked)}
                />
                <Sun className="swap-off" />
                <Moon className="swap-on" />
              </label>

              {/* 设置选项 */}
              <SettingsModal />

              {/* 回主页 */}
              <div className="tooltip tooltip-bottom" data-tip="媒体库主页">
                <button
                  className="btn btn-ghost btn-circle"
                  onClick={navigateToMediaServerHome}
                >
                  <House />
                </button>
              </div>
            </div>
          </div>

          {/* 折叠导航栏 */}
          <div className="collapse-content lg:hidden z-1">
            <ul className="menu">
              {navData.map((item) => navItemMapperVertical(item))}
            </ul>
          </div>
        </div>
      </nav>

      <main>
        <Outlet context={{ dark, setDark: setDarkAndSave }} />
      </main>
    </div>
  );
}
