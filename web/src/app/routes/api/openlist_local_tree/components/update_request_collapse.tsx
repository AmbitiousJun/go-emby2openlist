import { useEffect, useState } from "react";
import CommonCollapse from "~/components/settings_modal/common_collapse";

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

  // TODO: 开始更新按钮处理逻辑

  return (
    <CommonCollapse title={"手动更新目录树"} defaultChecked={true}>
      <div className="space-y-6">
        {/* 刷新前缀 */}
        <div className="flex items-center space-x-6 mt-6">
          <span className="font-bold text-base">路径前缀</span>
          <input
            type="text"
            className="input input-accent flex-1"
            placeholder="在此输入要更新的目录树路径前缀"
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
        <button className="btn btn-soft btn-accent">开始更新</button>
      </div>
    </CommonCollapse>
  );
}
