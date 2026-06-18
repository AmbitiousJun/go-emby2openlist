import { House, Menu } from "lucide-react";
import { Link, Outlet, useNavigate } from "react-router";
import { ModeToggle } from "~/components/mode_toggle";
import SettingsModal from "~/components/settings_modal/settings_modal";
import { Button } from "~/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
  navigationMenuTriggerStyle,
} from "~/components/ui/navigation-menu";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "~/components/ui/tooltip";
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

  const navigateToMediaServerHome = () => {
    window.location.href = `${window.location.origin}/`;
  };

  const navItemMapperHorizotal = (item: NavItem) => {
    if (item.children.length <= 0) {
      return (
        <NavigationMenuItem>
          <NavigationMenuLink asChild className={navigationMenuTriggerStyle()}>
            <Link to={item.to ?? "/"}>{item.label}</Link>
          </NavigationMenuLink>
        </NavigationMenuItem>
      );
    }
    return (
      <NavigationMenuItem>
        <NavigationMenuTrigger className={navigationMenuTriggerStyle()}>
          {item.label}
        </NavigationMenuTrigger>
        <NavigationMenuContent>
          {item.children.map((itemInner) => navItemMapperHorizotal(itemInner))}
        </NavigationMenuContent>
      </NavigationMenuItem>
    );
  };

  const navItemMapperVertical = (item: NavItem) => {
    if (item.children.length <= 0) {
      return (
        <DropdownMenuItem onClick={() => navigate(item.to ?? "/")}>
          {item.label}
        </DropdownMenuItem>
      );
    }

    return (
      <DropdownMenuSub>
        <DropdownMenuSubTrigger>{item.label}</DropdownMenuSubTrigger>
        <DropdownMenuPortal>
          <DropdownMenuSubContent>
            {item.children.map((itemInner) => navItemMapperVertical(itemInner))}
          </DropdownMenuSubContent>
        </DropdownMenuPortal>
      </DropdownMenuSub>
    );
  };

  return (
    <div>
      {/* 顶部信息栏 */}
      <div className="w-full flex justify-between items-center mb-12 shadow-sm border-b rounded-md px-4 py-4 bg-background text-backdround-foreground">
        {/* 左边的按钮 */}
        <div className="flex items-center">
          {/* 竖屏导航栏 使用下拉菜单代替 */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild className="lg:hidden">
              <Button variant="ghost" size="icon-lg">
                <Menu className="size-5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-40" align="start">
              {navData.map((nav) => navItemMapperVertical(nav))}
            </DropdownMenuContent>
          </DropdownMenu>

          <Button variant="ghost" className="text-xl">
            <Link to="/">Ge2o Web</Link>
          </Button>
        </div>

        {/* 中间的导航栏 */}
        <NavigationMenu className="hidden lg:block">
          <NavigationMenuList>
            {navData.map((nav) => navItemMapperHorizotal(nav))}
          </NavigationMenuList>
        </NavigationMenu>

        {/* 右边的按钮 */}
        <div className="flex items-center">
          {/* 切换主题 */}
          <ModeToggle />

          {/* 网站设置 */}
          <SettingsModal />

          {/* 回到服务器主页 */}
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon-lg"
                onClick={navigateToMediaServerHome}
              >
                <House className="size-5" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>回到媒体库主页</p>
            </TooltipContent>
          </Tooltip>
        </div>
      </div>

      <main>
        <Outlet context={{}} />
      </main>
    </div>
  );
}
