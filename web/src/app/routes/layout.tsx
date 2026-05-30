import { useEffect, useState } from "react";
import type { Route } from "./+types/layout";
import { Link, Outlet } from "react-router";
import { House, Menu, Moon, Sun, TvMinimalPlay } from "lucide-react";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Ge2o Web" },
    { name: "description", content: "go-emby2openlist web application" },
  ];
}

export default function Layout() {
  const [dark, setDark] = useState(localStorage.getItem("dark") ? true : false);

  useEffect(() => {
    document.documentElement.setAttribute(
      "data-theme",
      dark ? "dark" : "light",
    );
  }, [dark]);

  const setDarkAndSave = (dark: boolean) => {
    setDark(dark);
    localStorage.setItem("dark", dark ? "1" : "");
  };

  const navigateToMediaServerHome = () => {
    window.location.href = `${window.location.origin}/`;
  };

  return (
    <div>
      <nav>
        <div className="max-lg:collapse bg-base-200 mb-12 shadow-sm w-full rounded-md px-4">
          {/* 全屏遮罩 */}
          <input id="navbar-1-toggle" className="peer hidden" type="checkbox" />
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
              <button className="btn btn-ghost text-xl">Ge2o Web</button>
            </div>

            {/* 导航栏内容 */}
            <div className="navbar-center hidden lg:flex">
              <ul className="menu menu-horizontal px-1">
                <li>
                  <details>
                    <summary>接口调用</summary>
                    <ul className="p-2 bg-base-100 w-60 z-1">
                      <li>
                        <button>OpenList 本地目录树</button>
                      </li>
                    </ul>
                  </details>
                </li>
              </ul>
            </div>

            {/* 导航栏右侧 */}
            <div className="navbar-end space-x-4">
              {/* 主题切换 */}
              <label className="flex cursor-pointer gap-2">
                <Sun />
                <input
                  type="checkbox"
                  className="toggle"
                  checked={dark}
                  onChange={(e) => setDarkAndSave(e.target.checked)}
                />
                <Moon />
              </label>

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
              <li>
                <button>接口调用</button>
                <ul>
                  <li>
                    <button>OpenList 本地目录树</button>
                  </li>
                </ul>
              </li>
            </ul>
          </div>
        </div>
      </nav>

      <main>
        <Outlet />
      </main>
    </div>
  );
}
