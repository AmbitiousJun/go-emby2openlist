import { CircleAlert, CircleCheck, CircleX, Info } from "lucide-react";
import toast from "react-hot-toast";

// 弹出成功的消息提示
const success = (message: string) => {
  toast.custom(
    <div role="alert" className="alert alert-success alert-soft">
      <CircleCheck />
      <span>{message}</span>
    </div>,
  );
};

// 弹出普通消息提示
const info = (message: string) => {
  toast.custom(
    <div role="alert" className="alert alert-info alert-soft">
      <Info />
      <span>{message}</span>
    </div>,
  );
};

// 弹出警告消息提示
const warn = (message: string) => {
  toast.custom(
    <div role="alert" className="alert alert-warning alert-soft">
      <CircleAlert />
      <span>{message}</span>
    </div>,
  );
};

// 弹出错误消息提示
const error = (message: string) => {
  toast.custom(
    <div role="alert" className="alert alert-error alert-soft">
      <CircleX />
      <span>{message}</span>
    </div>,
  );
};

export default {
  success,
  info,
  warn,
  error,
};
