import { Settings } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import toastUtils from "../../utils/toast";

export default function SettingsModal() {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [apiSecret, setApiSecret] = useState(
    localStorage.getItem("api-secret") || "",
  );
  const [apiSecretChecking, setApiSecretChecking] = useState(false);
  const [apiSecretCheckOk, setApiSecretCheckOk] = useState<boolean | null>(
    null,
  );

  // 接口密钥变更的时候就自动检测正确性
  useEffect(() => {
    const timer = setTimeout(async () => {
      setApiSecretChecking(true);
      try {
        if (!apiSecret) {
          return;
        }

        const fetchState = await fetch("/ge2o/secret/validate", {
          method: "POST",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify({
            "secret": apiSecret
          })
        })

        const res = await fetchState.json();

        setApiSecretCheckOk(res.success ?? false);
      } catch (err) {
        dialogRef.current?.close();
        toastUtils.error(`校验接口密钥出现异常: ${err}`);
      } finally {
        setApiSecretChecking(false);
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [apiSecret]);

  // 显示对话框
  const show = () => {
    setApiSecret(localStorage.getItem("api-secret") || "");
    dialogRef.current?.showModal();
  };

  const saveAndClose = () => {
    localStorage.setItem("api-secret", apiSecret);
    dialogRef.current?.close();
    toastUtils.success("保存成功");
  };

  let apiSecretInputColor = "";
  if (apiSecretCheckOk === true) {
    apiSecretInputColor = "input-success";
  } else if (apiSecretCheckOk === false) {
    apiSecretInputColor = "input-error";
  }

  return (
    <>
      <button className="btn btn-ghost btn-circle" onClick={show}>
        <Settings />
      </button>
      <dialog ref={dialogRef} className="modal modal-bottom sm:modal-middle">
        <div className="modal-box">
          <h3 className="font-bold text-lg pb-4">设置选项</h3>

          <label className={`input w-full ${apiSecretInputColor}`}>
            接口密钥
            <input
              type="text"
              className="grow"
              placeholder="在此输入 config.yaml 中配置的程序密钥"
              value={apiSecret}
              onChange={(e) => setApiSecret(e.target.value)}
            />
            {apiSecretChecking && (
              <span className="loading loading-dots loading-xs"></span>
            )}
          </label>

          <div className="modal-action">
            <form method="dialog">
              {/* if there is a button in form, it will close the modal */}
              <button className="btn">关闭</button>
            </form>
            <button className="btn btn-accent" onClick={saveAndClose}>
              保存
            </button>
          </div>
        </div>
      </dialog>
    </>
  );
}
