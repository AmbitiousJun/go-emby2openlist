import { Settings } from "lucide-react";
import { useEffect, useRef, useState } from "react";

export default function SettingsModal() {
  const dialogRef = useRef<HTMLDialogElement>(null);
  const [apiSecret, setApiSecret] = useState(
    localStorage.getItem("api-secret") || "",
  );
  const [apiSecretChecking, setApiSecretChecking] = useState(false);
  const [apiSecretCheckOk, setApiSecretCheckOk] = useState<boolean | null>(null); 

  // 接口密钥变更的时候就自动检测正确性
  useEffect(() => {
    const timer = setTimeout(async () => {
      setApiSecretChecking(true);
      try {
        // TODO: 后端增加一个校验接口密钥的接口 并补充前端缺失的逻辑
        setApiSecretCheckOk(true);
      } finally {
        setApiSecretChecking(false);
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [apiSecret]);

  const saveAndClose = () => {
    localStorage.setItem("api-secret", apiSecret);
    dialogRef.current?.close();
  };

  let apiSecretInputColor = "";
  if (apiSecretCheckOk === true) {
    apiSecretInputColor = "input-success";
  } else if (apiSecretCheckOk === false) {
    apiSecretInputColor = "input-error";
  }

  return (
    <>
      <button
        className="btn btn-ghost btn-circle"
        onClick={() => dialogRef.current?.showModal()}
      >
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
