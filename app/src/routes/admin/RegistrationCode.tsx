import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card.tsx";
import { Button } from "@/components/ui/button.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Label } from "@/components/ui/label.tsx";
import { Badge } from "@/components/ui/badge.tsx";
import { NumberInput } from "@/components/ui/number-input.tsx";
import { Switch } from "@/components/ui/switch.tsx";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table.tsx";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog.tsx";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu.tsx";
import {
  Plus,
  MoreHorizontal,
  Trash2,
  Ban,
  Copy,
  Loader2,
  Ticket,
} from "lucide-react";
import { toast } from "sonner";
import {
  RegistrationCode,
  listRegistrationCodes,
  generateRegistrationCodes,
  getRegistrationCodeState,
  setRegistrationCodeState,
  disableRegistrationCode,
  deleteRegistrationCode,
} from "@/admin/api/registration-code.ts";
import { copyClipboard } from "@/utils/dom.ts";

// Admin registration code page.
//
// Visual style matches the other admin pages: single `<Card>` with a
// plain `CardHeader` + `CardTitle`, and `CardContent` containing a
// toolbar (toggle + refresh + create) above the table.

function RegistrationCodePage() {
  const { t } = useTranslation();
  const [codes, setCodes] = useState<RegistrationCode[]>([]);
  const [loading, setLoading] = useState(false);
  const [required, setRequired] = useState(false);
  const [genOpen, setGenOpen] = useState(false);

  async function refresh() {
    setLoading(true);
    const [listResp, stateResp] = await Promise.all([
      listRegistrationCodes(),
      getRegistrationCodeState(),
    ]);
    if (listResp.status && listResp.data) {
      setCodes(listResp.data);
    } else if (listResp.error) {
      toast.error(listResp.error);
    }
    setRequired(stateResp.required ?? false);
    setLoading(false);
  }

  useEffect(() => {
    refresh();
  }, []);

  async function toggleRequired(value: boolean) {
    setRequired(value);
    const resp = await setRegistrationCodeState(value);
    if (resp.status) {
      toast.success(
        value
          ? t("admin.reg-code.enabled")
          : t("admin.reg-code.disabled"),
      );
    } else {
      toast.error(resp.error || "Failed to update");
      setRequired(!value); // rollback
    }
  }

  async function copyCode(code: string) {
    await copyClipboard(code);
    toast.success(t("admin.reg-code.copied"));
  }

  async function remove(id: number) {
    if (!confirm(t("admin.reg-code.confirm-delete"))) return;
    const resp = await deleteRegistrationCode(id);
    if (resp.status) {
      toast.success(t("admin.reg-code.deleted") || "Deleted");
      refresh();
    } else {
      toast.error(resp.error || "Failed");
    }
  }

  async function disable(id: number) {
    const resp = await disableRegistrationCode(id);
    if (resp.status) {
      toast.success(t("admin.reg-code.disabled-action") || "Disabled");
      refresh();
    } else {
      toast.error(resp.error || "Failed");
    }
  }

  return (
    <div className={`reg-code`}>
      <Card className={`admin-card reg-code-card`}>
        <CardHeader className={`select-none`}>
          <CardTitle>{t("admin.reg-code.title")}</CardTitle>
        </CardHeader>
        <CardContent>
          {/* Toolbar */}
          <div className="flex items-center justify-between mb-4 pb-4 border-b">
            <div className="flex items-center gap-3">
              <Switch checked={required} onCheckedChange={toggleRequired} />
              <div>
                <Label className="font-medium cursor-pointer">
                  {t("admin.reg-code.require-label")}
                </Label>
                <p className="text-xs text-muted-foreground mt-0.5">
                  {t("admin.reg-code.require-desc")}
                </p>
              </div>
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={refresh}
                disabled={loading}
              >
                <Loader2
                  className={`h-4 w-4 mr-1.5 ${loading ? "animate-spin" : ""}`}
                />
                {t("refresh") || "Refresh"}
              </Button>
              <Button size="sm" onClick={() => setGenOpen(true)}>
                <Plus className="h-4 w-4 mr-1.5" />
                {t("admin.reg-code.generate")}
              </Button>
            </div>
          </div>

          {loading && codes.length === 0 ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : codes.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <Ticket className="w-10 h-10 mx-auto mb-3 opacity-50" />
              <p>{t("admin.reg-code.no-codes")}</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-16">#</TableHead>
                  <TableHead>{t("admin.reg-code.code")}</TableHead>
                  <TableHead className="w-24">
                    {t("admin.reg-code.uses")}
                  </TableHead>
                  <TableHead className="w-24">
                    {t("admin.reg-code.quota")}
                  </TableHead>
                  <TableHead>{t("admin.reg-code.note")}</TableHead>
                  <TableHead className="w-24">
                    {t("admin.reg-code.status")}
                  </TableHead>
                  <TableHead className="w-12"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {codes.map((c) => (
                  <TableRow key={c.id}>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      {c.id}
                    </TableCell>
                    <TableCell>
                      <button
                        className="font-mono text-xs hover:underline cursor-pointer"
                        onClick={() => copyCode(c.code)}
                      >
                        {c.code}
                      </button>
                    </TableCell>
                    <TableCell className="text-xs">
                      {c.used_count} / {c.max_uses}
                    </TableCell>
                    <TableCell className="text-xs">
                      {c.quota > 0 ? c.quota : "—"}
                    </TableCell>
                    <TableCell className="text-xs text-muted-foreground max-w-xs truncate">
                      {c.note || "—"}
                    </TableCell>
                    <TableCell>
                      {c.disabled ? (
                        <Badge variant="destructive">
                          {t("admin.reg-code.badge-disabled") || "Disabled"}
                        </Badge>
                      ) : c.used_count >= c.max_uses ? (
                        <Badge variant="secondary">
                          {t("admin.reg-code.badge-used") || "Used up"}
                        </Badge>
                      ) : (
                        <Badge variant="default">
                          {t("admin.reg-code.badge-active") || "Active"}
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                          >
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => copyCode(c.code)}>
                            <Copy className="h-4 w-4 mr-2" />
                            {t("copy") || "Copy"}
                          </DropdownMenuItem>
                          {!c.disabled && (
                            <DropdownMenuItem onClick={() => disable(c.id)}>
                              <Ban className="h-4 w-4 mr-2" />
                              {t("admin.reg-code.disable")}
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuItem
                            onClick={() => remove(c.id)}
                            className="text-destructive"
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            {t("admin.reg-code.delete")}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <GenerateDialog
        open={genOpen}
        onOpenChange={setGenOpen}
        onGenerated={() => {
          setGenOpen(false);
          refresh();
        }}
      />
    </div>
  );
}

function GenerateDialog(props: {
  open: boolean;
  onOpenChange: (o: boolean) => void;
  onGenerated: () => void;
}) {
  const { open, onOpenChange, onGenerated } = props;
  const { t } = useTranslation();
  const [number, setNumber] = useState(1);
  const [quota, setQuota] = useState(0);
  const [maxUses, setMaxUses] = useState(1);
  const [note, setNote] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit() {
    setLoading(true);
    const resp = await generateRegistrationCodes({
      number,
      quota,
      max_uses: maxUses,
      note,
    });
    setLoading(false);
    if (resp.status && resp.data) {
      toast.success(
        t("admin.reg-code.generated", { count: resp.data.length }) ||
          `${resp.data.length} codes generated`,
      );
      onGenerated();
    } else {
      toast.error(resp.error || "Failed");
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("admin.reg-code.generate-title")}</DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          <div>
            <Label>{t("admin.reg-code.number")}</Label>
            <NumberInput
              value={number}
              min={1}
              max={1000}
              onValueChange={(v) => setNumber(v as number)}
            />
          </div>
          <div>
            <Label>{t("admin.reg-code.max-uses")}</Label>
            <NumberInput
              value={maxUses}
              min={1}
              max={100000}
              onValueChange={(v) => setMaxUses(v as number)}
            />
            <p className="text-xs text-muted-foreground mt-1">
              {t("admin.reg-code.max-uses-tip")}
            </p>
          </div>
          <div>
            <Label>{t("admin.reg-code.bonus-quota")}</Label>
            <NumberInput
              value={quota}
              min={0}
              onValueChange={(v) => setQuota(v as number)}
            />
            <p className="text-xs text-muted-foreground mt-1">
              {t("admin.reg-code.bonus-quota-tip")}
            </p>
          </div>
          <div>
            <Label>{t("admin.reg-code.note")}</Label>
            <Input
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder={t("admin.reg-code.note-placeholder") || ""}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("cancel")}
          </Button>
          <Button onClick={submit} disabled={loading}>
            {loading && <Loader2 className="h-4 w-4 animate-spin mr-1" />}
            {t("confirm")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default RegistrationCodePage;
