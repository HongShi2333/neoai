import { useEffectAsync } from "@/utils/hook.ts";
import { useTranslation } from "react-i18next";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import {
  CommunityChannel,
  ChannelForm,
  listChannelsAdmin,
  createChannel,
  updateChannel,
  deleteChannel,
} from "@/api/community.ts";
import { withNotify } from "@/api/common.ts";
import { Button } from "@/components/ui/button.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Textarea } from "@/components/ui/textarea.tsx";
import { NumberInput } from "@/components/ui/number-input.tsx";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select.tsx";
import { MultiCombobox } from "@/components/ui/multi-combobox.tsx";
import {
  Table,
  TableBody,
  TableCell,
  TableHeader,
  TableRow,
} from "@/components/ui/table.tsx";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog.tsx";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog.tsx";
import OperationAction from "@/components/OperationAction.tsx";
import { Badge } from "@/components/ui/badge.tsx";
import {
  Hash,
  PencilLine,
  Plus,
  RotateCw,
  Trash,
  Lock,
  Globe,
} from "lucide-react";
import { allGroups } from "@/utils/groups.ts";

const emptyForm: ChannelForm = {
  name: "",
  description: "",
  topic: "",
  visibility: "public",
  visible_groups: [],
  post_groups: [],
  members: [],
  posters: [],
  position: 0,
};

// split a textarea string into a clean string list (comma or newline separated)
function parseList(raw: string): string[] {
  return raw
    .split(/[\n,]+/)
    .map((s) => s.trim())
    .filter((s) => s.length > 0);
}

function joinList(list: string[] | undefined): string {
  return (list || []).join("\n");
}

function toForm(ch: CommunityChannel): ChannelForm {
  return {
    name: ch.name,
    description: ch.description,
    topic: ch.topic,
    visibility: ch.visibility,
    visible_groups: ch.visible_groups || [],
    post_groups: ch.post_groups || [],
    members: ch.members || [],
    posters: ch.posters || [],
    position: ch.position,
  };
}

type EditorDialogProps = {
  open: boolean;
  setOpen: (open: boolean) => void;
  editing: CommunityChannel | null;
  onRefresh: () => void;
};

function EditorDialog({
  open,
  setOpen,
  editing,
  onRefresh,
}: EditorDialogProps) {
  const { t } = useTranslation();
  const [form, setForm] = useState<ChannelForm>(emptyForm);
  const [membersRaw, setMembersRaw] = useState("");
  const [postersRaw, setPostersRaw] = useState("");
  const [_, setLoading] = useState(false);

  useEffectAsync(async () => {
    if (!open) return;
    if (editing) {
      const f = toForm(editing);
      setForm(f);
      setMembersRaw(joinList(f.members));
      setPostersRaw(joinList(f.posters));
    } else {
      setForm(emptyForm);
      setMembersRaw("");
      setPostersRaw("");
    }
  }, [open, editing]);

  const set = <K extends keyof ChannelForm>(key: K, value: ChannelForm[K]) =>
    setForm((f) => ({ ...f, [key]: value }));

  async function post() {
    const payload: ChannelForm = {
      ...form,
      name: form.name.trim(),
      members: form.visibility === "private" ? parseList(membersRaw) : [],
      posters: parseList(postersRaw),
    };
    if (payload.name === "") {
      toast.error(t("error"), { description: t("admin.community.name-required") });
      return;
    }
    setLoading(true);
    const resp = editing
      ? await updateChannel(editing.id, payload)
      : await createChannel(payload);
    setLoading(false);
    withNotify(t, resp, true);
    if (resp.status) {
      setOpen(false);
      onRefresh();
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className={`max-h-[90vh] overflow-y-auto`}>
        <DialogHeader>
          <DialogTitle>
            {editing
              ? t("admin.community.edit")
              : t("admin.community.create")}
          </DialogTitle>
        </DialogHeader>

        <div className={`flex flex-col gap-4 py-2`}>
          <div className={`flex flex-col gap-1.5`}>
            <label className={`text-sm font-medium`}>
              {t("admin.community.name")}
            </label>
            <Input
              value={form.name}
              placeholder={t("admin.community.name-placeholder")}
              onChange={(e) => set("name", e.target.value)}
            />
          </div>

          <div className={`flex flex-col gap-1.5`}>
            <label className={`text-sm font-medium`}>
              {t("admin.community.topic")}
            </label>
            <Input
              value={form.topic}
              placeholder={t("admin.community.topic-placeholder")}
              onChange={(e) => set("topic", e.target.value)}
            />
          </div>

          <div className={`flex flex-col gap-1.5`}>
            <label className={`text-sm font-medium`}>
              {t("admin.community.description")}
            </label>
            <Textarea
              rows={2}
              value={form.description}
              placeholder={t("admin.community.description-placeholder")}
              onChange={(e) => set("description", e.target.value)}
            />
          </div>

          <div className={`flex flex-row gap-4`}>
            <div className={`flex flex-col gap-1.5 grow`}>
              <label className={`text-sm font-medium`}>
                {t("admin.community.visibility")}
              </label>
              <Select
                value={form.visibility}
                onValueChange={(v) =>
                  set("visibility", v as "public" | "private")
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="public">
                    {t("admin.community.public")}
                  </SelectItem>
                  <SelectItem value="private">
                    {t("admin.community.private")}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className={`flex flex-col gap-1.5 w-32`}>
              <label className={`text-sm font-medium`}>
                {t("admin.community.position")}
              </label>
              <NumberInput
                value={form.position}
                acceptNegative
                onValueChange={(v) => set("position", v)}
              />
            </div>
          </div>

          {form.visibility === "public" && (
            <div className={`flex flex-col gap-1.5`}>
              <label className={`text-sm font-medium`}>
                {t("admin.community.visible-groups")}
              </label>
              <MultiCombobox
                className={`w-full max-w-full`}
                value={form.visible_groups}
                align={`end`}
                onChange={(v) => set("visible_groups", v)}
                list={allGroups}
                listTranslate={"admin.channels.groups"}
                placeholder={t("admin.community.visible-groups-placeholder", {
                  length: form.visible_groups.length,
                })}
              />
              <p className={`text-xs text-secondary`}>
                {t("admin.community.visible-groups-tip")}
              </p>
            </div>
          )}

          {form.visibility === "private" && (
            <div className={`flex flex-col gap-1.5`}>
              <label className={`text-sm font-medium`}>
                {t("admin.community.members")}
              </label>
              <Textarea
                rows={3}
                value={membersRaw}
                placeholder={t("admin.community.members-placeholder")}
                onChange={(e) => setMembersRaw(e.target.value)}
              />
              <p className={`text-xs text-secondary`}>
                {t("admin.community.members-tip")}
              </p>
            </div>
          )}

          <div className={`flex flex-col gap-1.5`}>
            <label className={`text-sm font-medium`}>
              {t("admin.community.post-groups")}
            </label>
            <MultiCombobox
              className={`w-full max-w-full`}
              value={form.post_groups}
              align={`end`}
              onChange={(v) => set("post_groups", v)}
              list={allGroups}
              listTranslate={"admin.channels.groups"}
              placeholder={t("admin.community.post-groups-placeholder", {
                length: form.post_groups.length,
              })}
            />
            <p className={`text-xs text-secondary`}>
              {t("admin.community.post-groups-tip")}
            </p>
          </div>

          <div className={`flex flex-col gap-1.5`}>
            <label className={`text-sm font-medium`}>
              {t("admin.community.posters")}
            </label>
            <Textarea
              rows={3}
              value={postersRaw}
              placeholder={t("admin.community.posters-placeholder")}
              onChange={(e) => setPostersRaw(e.target.value)}
            />
            <p className={`text-xs text-secondary`}>
              {t("admin.community.posters-tip")}
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button variant={`outline`} onClick={() => setOpen(false)}>
            {t("cancel")}
          </Button>
          <Button loading={true} onLoadingChange={setLoading} onClick={post}>
            {t("confirm")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function CommunityTable() {
  const { t } = useTranslation();
  const [data, setData] = useState<CommunityChannel[]>([]);
  const [loading, setLoading] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editing, setEditing] = useState<CommunityChannel | null>(null);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  async function refresh() {
    setLoading(true);
    const resp = await listChannelsAdmin();
    setLoading(false);
    withNotify(t, resp);
    setData(resp.data || []);
  }

  useEffectAsync(async () => {
    await refresh();
  }, []);

  const deleting = useMemo(
    () => data.find((c) => c.id === deleteId) || null,
    [data, deleteId],
  );

  return (
    <div className={`community-table`}>
      <div className={`flex flex-row w-full h-max mb-4 gap-2`}>
        <Button
          onClick={() => {
            setEditing(null);
            setEditorOpen(true);
          }}
        >
          <Plus className={`w-4 h-4 mr-2`} />
          {t("admin.community.create")}
        </Button>
        <div className={`grow`} />
        <Button variant={`outline`} size={`icon`} onClick={refresh}>
          <RotateCw className={`w-4 h-4 ${loading ? "animate-spin" : ""}`} />
        </Button>
      </div>

      <Table classNameWrapper={`table`}>
        <TableHeader>
          <TableRow className={`select-none whitespace-nowrap`}>
            <TableCell>{t("admin.community.id")}</TableCell>
            <TableCell>{t("admin.community.name")}</TableCell>
            <TableCell>{t("admin.community.visibility")}</TableCell>
            <TableCell>{t("admin.community.topic")}</TableCell>
            <TableCell>{t("admin.community.position")}</TableCell>
            <TableCell>{t("admin.community.created-by")}</TableCell>
            <TableCell>{t("admin.community.action")}</TableCell>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((ch) => (
            <TableRow key={ch.id}>
              <TableCell>{ch.id}</TableCell>
              <TableCell>
                <div className={`flex flex-row items-center gap-1.5`}>
                  <Hash className={`w-3.5 h-3.5`} />
                  <span className={`font-medium`}>{ch.name}</span>
                </div>
              </TableCell>
              <TableCell>
                <Badge variant={ch.visibility === "private" ? `destructive` : `secondary`}>
                  {ch.visibility === "private" ? (
                    <Lock className={`w-3 h-3 mr-1`} />
                  ) : (
                    <Globe className={`w-3 h-3 mr-1`} />
                  )}
                  {t(`admin.community.${ch.visibility}`)}
                </Badge>
              </TableCell>
              <TableCell className={`text-secondary`}>
                {ch.topic || "-"}
              </TableCell>
              <TableCell>{ch.position}</TableCell>
              <TableCell className={`text-secondary`}>
                {ch.created_by || "-"}
              </TableCell>
              <TableCell>
                <div className={`inline-flex flex-row gap-2`}>
                  <OperationAction
                    tooltip={t("admin.community.edit")}
                    onClick={() => {
                      setEditing(ch);
                      setEditorOpen(true);
                    }}
                  >
                    <PencilLine className={`h-4 w-4`} />
                  </OperationAction>
                  <OperationAction
                    tooltip={t("admin.community.delete")}
                    variant={`destructive`}
                    onClick={() => setDeleteId(ch.id)}
                  >
                    <Trash className={`h-4 w-4`} />
                  </OperationAction>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      <EditorDialog
        open={editorOpen}
        setOpen={setEditorOpen}
        editing={editing}
        onRefresh={refresh}
      />

      <AlertDialog open={deleteId !== null} onOpenChange={(o) => !o && setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {t("admin.community.delete")}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {t("admin.community.delete-confirm", {
                name: deleting?.name || "",
              })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("cancel")}</AlertDialogCancel>
            <AlertDialogAction
              onClick={async () => {
                if (deleteId === null) return;
                const resp = await deleteChannel(deleteId);
                withNotify(t, resp, true);
                setDeleteId(null);
                refresh();
              }}
            >
              {t("confirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

export default CommunityTable;
