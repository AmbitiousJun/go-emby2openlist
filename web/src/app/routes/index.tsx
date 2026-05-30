import logoImg from "~/assets/logo.png";
import { FaGithub } from "react-icons/fa";
import { Cog } from "lucide-react";

export default function Index() {
  return (
    <div className="w-full lg:px-48 max-lg:px-12 space-y-6">
      <img className="w-50 h-50 mx-auto" src={logoImg} alt="logo" />

      <div className="text-4xl font-bold text-center">go-emby2openlist</div>

      <div className="flex flex-wrap gap-2 items-center justify-center">
        <a href="https://github.com/AmbitiousJun/go-emby2openlist/tree/v2.7.1" target="_blank">
          <img src="https://img.shields.io/github/v/tag/AmbitiousJun/go-emby2openlist"></img>
        </a>
        <a href="https://hub.docker.com/r/ambitiousjun/go-emby2openlist/tags" target="_blank">
          <img src="https://img.shields.io/docker/image-size/ambitiousjun/go-emby2openlist/v2.7.1"></img>
        </a>
        <a href="https://hub.docker.com/r/ambitiousjun/go-emby2openlist/tags" target="_blank">
          <img src="https://img.shields.io/docker/pulls/ambitiousjun/go-emby2openlist"></img>
        </a>
        <a href="https://github.com/AmbitiousJun/go-emby2openlist/releases/latest" target="_blank">
          <img src="https://img.shields.io/github/downloads/AmbitiousJun/go-emby2openlist/total"></img>
        </a>
        <a href="https://goreportcard.com/report/github.com/AmbitiousJun/go-emby2openlist/v2" target="_blank">
          <img src="https://goreportcard.com/badge/github.com/AmbitiousJun/go-emby2openlist/v2"></img>
        </a>
        <img src="https://img.shields.io/github/stars/AmbitiousJun/go-emby2openlist"></img>
        <img src="https://img.shields.io/github/license/AmbitiousJun/go-emby2openlist"></img>
      </div>

      <div className="text-center text-lg">
        Go 语言编写的 Emby + OpenList
        网盘直链反向代理服务，深度适配阿里云盘转码播放。
      </div>

      <div className="max-w-[300px] w-full space-y-6 px-4 mx-auto">
        <nav className="rounded-3xl border border-gray-200 p-6 dark:border-gray-700 space-y-4">
          <p className="leading-6 text-center">
            快速导航
          </p>
          <ul>
            {resources.map(({ href, text, icon }) => (
              <li key={href}>
                <a
                  className="group flex items-center gap-3 self-stretch p-3 leading-normal text-blue-700 hover:underline dark:text-blue-500"
                  href={href}
                  target="_blank"
                  rel="noreferrer"
                >
                  {icon}
                  {text}
                </a>
              </li>
            ))}
          </ul>
        </nav>
      </div>
    </div>
  );
}

const resources = [
  {
    href: "https://github.com/AmbitiousJun/go-emby2openlist",
    text: "Github",
    icon: <FaGithub size={24} />,
  },
  {
    href: "https://github.com/AmbitiousJun/go-emby2openlist/blob/v2.7.1/config-example.yml",
    text: "示例配置",
    icon: <Cog />,
  },
];
