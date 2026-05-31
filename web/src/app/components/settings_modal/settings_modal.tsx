import { Settings } from "lucide-react";
import { useRef } from "react";

export default function SettingsModal() {
  const dialogRef = useRef<HTMLDialogElement>(null);

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

          <label className="input w-full">
            接口密钥
            <input type="text" className="grow" placeholder="在此输入 config.yaml 中配置的程序密钥" />
          </label>

          <div className="modal-action">
            <form method="dialog">
              {/* if there is a button in form, it will close the modal */}
              <button className="btn">关闭</button>
            </form>
            <button className="btn btn-accent">保存</button>
          </div>
        </div>
      </dialog>
    </>
  );
}
