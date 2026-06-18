import { Settings } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { Button } from "~/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/components/ui/dialog";
import { DropdownMenuSeparator } from "~/components/ui/dropdown-menu";
import { Field, FieldGroup } from "~/components/ui/field";
import {
  InputGroup,
  InputGroupAddon,
  InputGroupInput,
  InputGroupText,
} from "~/components/ui/input-group";
import { Spinner } from "~/components/ui/spinner";

export const LOCAL_STORAGE_KEY_API_SECRET = "api_secret";

export default function SettingsModal() {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [apiSecret, setApiSecret] = useState(
    localStorage.getItem(LOCAL_STORAGE_KEY_API_SECRET) || "",
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
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            secret: apiSecret,
          }),
        });

        if (!fetchState.ok || fetchState.status != 200) {
          throw Error(`请求失败: ${fetchState.statusText}`);
        }
        const res = await fetchState.json();

        setApiSecretCheckOk(res.success ?? false);
      } catch (err) {
        if (err instanceof Error) {
          err = err.message;
        }

        setDialogOpen(false);
        toast.error(`校验接口密钥异常: ${err}`);
      } finally {
        setApiSecretChecking(false);
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [apiSecret]);

  // 对话框打开的时候自动加载缓存中的最新数据
  useEffect(() => {
    if (!dialogOpen) {
      return;
    }
    setApiSecret(localStorage.getItem(LOCAL_STORAGE_KEY_API_SECRET) || "");
  }, [dialogOpen]);

  // 保存并关闭对话框
  const saveAndClose = () => {
    localStorage.setItem(LOCAL_STORAGE_KEY_API_SECRET, apiSecret);
    setDialogOpen(false);
    toast.success("保存成功");
  };

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button variant="ghost">
          <Settings className="size-5" />
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            saveAndClose();
          }}
        >
          <DialogHeader>
            <DialogTitle className="text-lg">设置选项</DialogTitle>
          </DialogHeader>

          <DropdownMenuSeparator className="mb-4" />

          <FieldGroup>
            <Field>
              <InputGroup>
                <InputGroupInput
                  id="api-secret"
                  placeholder="在此输入 config.yaml 中配置的程序密钥"
                  aria-invalid={!apiSecretCheckOk}
                  value={apiSecret}
                  onChange={(e) => setApiSecret(e.target.value)}
                />
                <InputGroupAddon>
                  <InputGroupText>接口密钥</InputGroupText>
                </InputGroupAddon>
                {apiSecretChecking && (
                  <InputGroupAddon align="inline-end">
                    <Spinner />
                  </InputGroupAddon>
                )}
              </InputGroup>
            </Field>
          </FieldGroup>

          <DialogFooter>
            <DialogClose asChild>
              <Button variant="outline">关 闭</Button>
            </DialogClose>
            <Button type="submit">保 存</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
