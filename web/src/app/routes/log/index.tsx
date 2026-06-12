import Convert from "ansi-to-html";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { LOCAL_STORAGE_KEY_API_SECRET } from "~/components/settings_modal/settings_modal";

const convert = new Convert();

export default function Log() {
  const logRef = useRef<HTMLDivElement>(null);
  const isAutoScrollingRef = useRef(false);
  const [logs, setLogs] = useState<string[]>([]);
  const [autoScrollToBottom, setAutoScrollToBottom] = useState(true);

  // 路由加载完成时自动建立 WebSocket 连接读取日志
  useEffect(() => {
    // 1 校验密钥
    const secret = localStorage.getItem(LOCAL_STORAGE_KEY_API_SECRET);
    if (!secret) {
      toast.info("请先配置接口密钥");
      return;
    }

    // 2 拼接地址
    let prefix = "ws://";
    if (location.protocol.startsWith("https")) {
      prefix = "wss://";
    }
    const wsUrl = `${prefix}${location.host}/ge2o/ws/log/sync?secret=${secret}`;

    // 3 建立连接
    appendLog("正在尝试建立连接...");
    const ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      appendLog(convert.toHtml("\x1b[32m成功建立连接 😄\x1b[0m"));
    };

    ws.onerror = () => {
      appendLog(
        convert.toHtml(
          `\x1b[38;2;228;116;112m连接建立失败！请确保接口密钥配置正确\x1b[0m`,
        ),
      );
    };

    ws.onclose = () => {
      appendLog(convert.toHtml("\x1b[38;2;145;147;152m连接已关闭\x1b[0m"));
    };

    ws.onmessage = (e) => {
      appendLog(convert.toHtml(e.data));
    };

    return () => ws.close();
  }, []);

  // 监听日志变化自动滚动到底部
  useEffect(() => {
    const el = logRef.current;
    if (el && autoScrollToBottom) {
      scrollToBottom(el);
    }
  }, [logs]);

  const scrollToBottom = (el: HTMLDivElement) => {
    isAutoScrollingRef.current = true;

    el.scrollTop = el.scrollHeight;

    requestAnimationFrame(() => {
      isAutoScrollingRef.current = false;
    });
  };

  // 追加日志
  const appendLog = (newLog: string) => {
    setLogs((prev) => {
      let arr = [...prev, newLog];

      if (arr.length > 5000) {
        arr = arr.slice(arr.length - 5000);
      }

      return arr;
    });
  };

  const handleScroll = (el: HTMLDivElement) => {
    if (isAutoScrollingRef.current) {
      return;
    }

    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 20;
    if (atBottom) {
      return;
    }

    // 用户手动滚动 关闭自动跟踪
    if (autoScrollToBottom) {
      setAutoScrollToBottom(false);
    }
  };

  return (
    <div className="w-full lg:px-48 max-lg:px-12 space-y-6 pb-12">
      <fieldset className="fieldset bg-base-100 border-base-300 rounded-box w-64 border p-4">
        <legend className="fieldset-legend">终端选项</legend>
        <label className="label">
          <input
            type="checkbox"
            className="toggle toggle-accent"
            checked={autoScrollToBottom}
            onChange={(e) => setAutoScrollToBottom(e.target.checked)}
          />
          自动滚动到底部
        </label>
      </fieldset>

      <div
        ref={logRef}
        className="h-[calc(100vh-18rem)] overflow-auto rounded-lg bg-neutral text-neutral-content p-4 font-mono text-sm whitespace-pre-wrap"
        onScroll={(e) => handleScroll(e.currentTarget)}
      >
        {logs.map((log, i) => (
          <div
            key={i}
            dangerouslySetInnerHTML={{
              __html: log,
            }}
          />
        ))}
      </div>
    </div>
  );
}
