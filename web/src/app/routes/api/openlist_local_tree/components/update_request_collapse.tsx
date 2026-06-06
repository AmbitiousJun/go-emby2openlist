import { useState } from "react";
import CommonCollapse from "~/components/settings_modal/common_collapse";
import { LOCAL_STORAGE_KEY_API_SECRET } from "~/components/settings_modal/settings_modal";
import toastUtils from "~/utils/toast";

const LOCAL_STORAGE_KEY_FORCE_REFRESH =
  "api:openlist_local_tree:update_request:force_request";
const LOCAL_STORAGE_KEY_UPDATE_PREFIX =
  "api:openlist_local_tree:update_request:update_prefix";
const LOCAL_STORAGE_KEY_PREFIX_HISTORIES =
  "api:openlist_local_tree:update_request:prefix_histories";

export default function UpdateRequestCollapse() {
  const [forceRefreshFlag, setForceRefreshFlag] = useState(
    localStorage.getItem(LOCAL_STORAGE_KEY_FORCE_REFRESH) ? true : false,
  );
  const [prefix, setPrefix] = useState(
    localStorage.getItem(LOCAL_STORAGE_KEY_UPDATE_PREFIX) ?? "",
  );
  const [prefixHistories, setPrefixHistories] = useState<string[]>(
    JSON.parse(
      localStorage.getItem(LOCAL_STORAGE_KEY_PREFIX_HISTORIES) ?? "[]",
    ),
  );
  const [updating, setUpdating] = useState(false);

  const updateForceRefreshFlagAndSave = (flag: boolean) => {
    setForceRefreshFlag(flag);
    if (flag) {
      localStorage.setItem(LOCAL_STORAGE_KEY_FORCE_REFRESH, "1");
    } else {
      localStorage.removeItem(LOCAL_STORAGE_KEY_FORCE_REFRESH);
    }
  };

  const updatePrefixAndSave = (prefix: string) => {
    setPrefix(prefix);
    localStorage.setItem(LOCAL_STORAGE_KEY_UPDATE_PREFIX, prefix);
  };

  // 调用接口触发后台目录树刷新
  const handleUpdate = async () => {
    setUpdating(true);
    try {
      // 1 校验密钥
      const secret = localStorage.getItem(LOCAL_STORAGE_KEY_API_SECRET);
      if (!secret) {
        toastUtils.info("请先设置接口密钥");
        return;
      }

      // 2 发起请求
      const fetchState = await fetch("/ge2o/openlist/local_tree/update", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          secret: secret,
          prefix: prefix.trim(),
          refresh: forceRefreshFlag,
        }),
      });

      // 3 响应解析
      if (!fetchState.ok || fetchState.status != 200) {
        throw Error(`请求失败: ${fetchState.statusText}`);
      }
      const res = (await fetchState.json()) as {
        success: boolean;
        message: string;
      };

      if (!res.success) {
        toastUtils.warn(res.message);
        return;
      }
      toastUtils.success(res.message);
    } catch (err) {
      if (err instanceof Error) {
        err = err.message;
      }
      toastUtils.error(`手动更新目录树异常: ${err}`);
    } finally {
      setUpdating(false);
    }
  };

  return (
    <CommonCollapse title={"手动更新目录树"} defaultChecked={true}>
      <div className="space-y-6">
        {/* 刷新前缀 */}
        <div className="flex items-center space-x-6 mt-6">
          <span className="font-bold text-base">路径前缀</span>
          <input
            type="text"
            className="input input-accent flex-1"
            placeholder="在此输入要更新的目录树路径前缀，不指定前缀时进行全量更新"
            list="prefix-histories-datalist"
            value={prefix}
            onChange={(e) => updatePrefixAndSave(e.target.value)}
          />
          {/* TODO: 等到接口调用通了之后再来完善这个功能 */}
          <datalist id="prefix-histories-datalist">
            <option value="Chrome"></option>
            <option value="Firefox"></option>
            <option value="Safari"></option>
            <option value="Opera"></option>
            <option value="Edge"></option>
          </datalist>
        </div>

        {/* 强制刷新 */}
        <div className="flex items-center space-x-6">
          <span className="font-bold text-base">强制刷新</span>
          <input
            type="checkbox"
            className="toggle toggle-accent"
            checked={forceRefreshFlag}
            onChange={(e) => updateForceRefreshFlagAndSave(e.target.checked)}
          />
        </div>

        {/* 更新按钮 */}
        <button
          className={`btn btn-soft btn-accent ${updating && "btn-disabled"}`}
          onClick={handleUpdate}
        >
          {updating && <span className="loading loading-spinner w-4 h-4" />}
          {updating ? "请稍候..." : "开始更新"}
        </button>
      </div>
    </CommonCollapse>
  );
}
