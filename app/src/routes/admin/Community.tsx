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
import { Textarea } from "@/components/ui/textarea.tsx";
import { Label } from "@/components/ui/label.tsx";
import { Badge } from "@/components/ui/badge.tsx";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select.tsx";
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
  Hash,
  Plus,
  Settings2,
  Trash2,
  MoreHorizontal,
  Globe,
  Users,
  Shield,
  Lock,
  Loader2,
} from "lucide-react";
import { toast } from "sonner";
import {
  Channel,
  ChannelForm,
  Visibility,
  SendPermission,
  createChannel,
  updateChannel,
  deleteChannel,
} from "@/api/community.ts";
import { adminListChannels } from "@/admin/api/community.ts";

// Admin community management page.
//
// Visual style matches the other admin pages (Broadcast / Users /
// Channel): a single `<Card>` with `CardHeader` containing only the
// title, and `CardContent` containing a toolbar (with the create
// button on the right) above the channel table.

const VIS_OPTIONS: { value: Visibility; label: string; icon: any }[] = [
  { value: "public", label: "Public", icon: Globe },
  { value: "members", label: "Members only", icon: Users },
  { value: "roles", label: "By role", icon: Shield },
  { value: "whitelist", label: "Whitelist", icon: Lock },
];

const SEND_OPTIONS: { value: SendPermission; label: string }[] = [
  { value: "everyone", label: "Everyone (who can see)" },
  { value: "admins", label: "Admins only" },
  { value: "whitelist", label: "Whitelisted senders" },
];

function Community() {
  const { t } = useTranslation();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(false);
  const [editorOpen, setEditorOpen] = useState(false);
  const [editing, setEditing] = useState<Channel | null>(null);

  async function refresh() {
    setLoading(true);
    const resp = await adminListChannels();
    if (resp.status && resp.data) {
      setChannels(resp.data);
    } else if (resp.error) {
      toast.error(resp.error);
    }
    setLoading(false);
  }

  useEffect(() => {
    refresh();
  }, []);

  function openCreate() {
    setEditing(null);
    setEditorOpen(true);
  }

  function openEdit(ch: Channel) {
    setEditing(ch);
    setEditorOpen(true);
  }

  async function submit(form: ChannelForm) {
    if (editing) {
      const resp = await updateChannel(editing.id, form);
      if (resp.status) {
        toast.success(t("admin.community.updated") || "Channel updated");
        setEditorOpen(false);
        refresh();
      } else {
        toast.error(resp.error || "update failed");
      }
    } else {
      const resp = await createChannel(form);
      if (resp.status) {
        toast.success(t("admin.community.created") || "Channel created");
        setEditorOpen(false);
        refresh();
      } else {
        toast.error(resp.error || "create failed");
      }
    }
  }

  async function remove(id: number) {
    if (!confirm(t("admin.community.confirm-delete"))) return;
    const resp = await deleteChannel(id);
    if (resp.status) {
      toast.success(t("admin.community.deleted") || "Channel deleted");
      refresh();
    } else {
      toast.error(resp.error || "delete failed");
    }
  }

  return (
    <div className={`community`}>
      <Card className={`admin-card community-card`}>
        <CardHeader className={`select-none`}>
          <CardTitle>{t("admin.community.title")}</CardTitle>
        </CardHeader>
        <CardContent>
          {/* Toolbar: refresh + create button. Buttons live INSIDE
              CardContent, not CardHeader, so they don't get pushed
              around by the title's wrapping behaviour on narrow
              screens. */}
          <div className="flex items-center justify-between mb-4">
            <p className="text-sm text-muted-foreground">
              {t("admin.community.subtitle") ||
                "Discord-style channels with visibility + send permissions."}
            </p>
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
              <Button size="sm" onClick={openCreate}>
                <Plus className="h-4 w-4 mr-1.5" />
                {t("admin.community.create")}
              </Button>
            </div>
          </div>

          {loading && channels.length === 0 ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : channels.length === 0 ? (
            <div className="text-center py-12 text-muted-foreground">
              <Hash className="w-10 h-10 mx-auto mb-3 opacity-50" />
              <p>{t("admin.community.no-channels")}</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">#</TableHead>
                  <TableHead>{t("admin.community.name")}</TableHead>
                  <TableHead>{t("admin.community.topic")}</TableHead>
                  <TableHead>{t("admin.community.visibility")}</TableHead>
                  <TableHead>
                    {t("admin.community.send-permission")}
                  </TableHead>
                  <TableHead className="w-16"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {channels.map((ch) => (
                  <TableRow key={ch.id}>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      {ch.id}
                    </TableCell>
                    <TableCell className="font-medium">
                      <div className="flex items-center gap-2">
                        <Hash className="h-3.5 w-3.5 text-muted-foreground" />
                        {ch.name}
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground max-w-xs truncate">
                      {ch.topic || "—"}
                    </TableCell>
                    <TableCell>
                      <VisibilityBadge v={ch.visibility} />
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {ch.send_permission}
                      </Badge>
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
                          <DropdownMenuItem onClick={() => openEdit(ch)}>
                            <Settings2 className="h-4 w-4 mr-2" />
                            {t("admin.community.edit")}
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => remove(ch.id)}
                            className="text-destructive"
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            {t("admin.community.delete")}
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

      <ChannelEditorDialog
        open={editorOpen}
        onOpenChange={setEditorOpen}
        channel={editing}
        onSubmit={submit}
      />
    </div>
  );
}

function VisibilityBadge({ v }: { v: Visibility }) {
  const opt = VIS_OPTIONS.find((o) => o.value === v);
  if (!opt) return null;
  const Icon = opt.icon;
  return (
    <Badge variant="outline" className="text-xs">
      <Icon className="h-3 w-3 mr-1" />
      {opt.label}
    </Badge>
  );
}

function ChannelEditorDialog(props: {
  open: boolean;
  onOpenChange: (o: boolean) => void;
  channel: Channel | null;
  onSubmit: (form: ChannelForm) => void;
}) {
  const { open, onOpenChange, channel, onSubmit } = props;
  const { t } = useTranslation();

  const [name, setName] = useState("");
  const [topic, setTopic] = useState("");
  const [visibility, setVisibility] = useState<Visibility>("members");
  const [sendPerm, setSendPerm] = useState<SendPermission>("everyone");
  const [visibleRoles, setVisibleRoles] = useState("");
  const [visibleUsers, setVisibleUsers] = useState("");
  const [senders, setSenders] = useState("");

  useEffect(() => {
    if (channel) {
      setName(channel.name);
      setTopic(channel.topic);
      setVisibility(channel.visibility);
      setSendPerm(channel.send_permission);
      setVisibleRoles((channel.visible_roles || []).join(","));
      setVisibleUsers((channel.visible_users || []).join(","));
      setSenders((channel.senders || []).join(","));
    } else {
      setName("");
      setTopic("");
      setVisibility("members");
      setSendPerm("everyone");
      setVisibleRoles("");
      setVisibleUsers("");
      setSenders("");
    }
  }, [channel, open]);

  function submit() {
    if (!name.trim()) {
      toast.error(t("admin.community.name-required") || "Name required");
      return;
    }
    onSubmit({
      name: name.trim(),
      topic: topic.trim(),
      visibility,
      send_permission: sendPerm,
      visible_roles: visibleRoles
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean),
      visible_users: visibleUsers
        .split(",")
        .map((s) => parseInt(s.trim(), 10))
        .filter((n) => !isNaN(n)),
      senders: senders
        .split(",")
        .map((s) => parseInt(s.trim(), 10))
        .filter((n) => !isNaN(n)),
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>
            {channel
              ? t("admin.community.edit-channel")
              : t("admin.community.create-channel")}
          </DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          <div>
            <Label>{t("admin.community.name")}</Label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="general"
            />
          </div>
          <div>
            <Label>{t("admin.community.topic")}</Label>
            <Textarea
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              placeholder={t("admin.community.topic-placeholder") || ""}
              rows={2}
            />
          </div>
          <div>
            <Label>{t("admin.community.visibility")}</Label>
            <Select
              value={visibility}
              onValueChange={(v) => setVisibility(v as Visibility)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {VIS_OPTIONS.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          {visibility === "roles" && (
            <div>
              <Label>{t("admin.community.visible-roles")}</Label>
              <Input
                value={visibleRoles}
                onChange={(e) => setVisibleRoles(e.target.value)}
                placeholder="normal,basic,standard,pro,admin"
              />
              <p className="text-xs text-muted-foreground mt-1">
                {t("admin.community.visible-roles-tip") ||
                  "Comma-separated group names"}
              </p>
            </div>
          )}
          {visibility === "whitelist" && (
            <div>
              <Label>{t("admin.community.visible-users")}</Label>
              <Input
                value={visibleUsers}
                onChange={(e) => setVisibleUsers(e.target.value)}
                placeholder="1,2,3"
              />
              <p className="text-xs text-muted-foreground mt-1">
                {t("admin.community.user-ids-tip") || "Comma-separated user IDs"}
              </p>
            </div>
          )}
          <div>
            <Label>{t("admin.community.send-permission")}</Label>
            <Select
              value={sendPerm}
              onValueChange={(v) => setSendPerm(v as SendPermission)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {SEND_OPTIONS.map((o) => (
                  <SelectItem key={o.value} value={o.value}>
                    {o.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          {sendPerm === "whitelist" && (
            <div>
              <Label>{t("admin.community.senders")}</Label>
              <Input
                value={senders}
                onChange={(e) => setSenders(e.target.value)}
                placeholder="1,2,3"
              />
              <p className="text-xs text-muted-foreground mt-1">
                {t("admin.community.user-ids-tip") || "Comma-separated user IDs"}
              </p>
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("cancel")}
          </Button>
          <Button onClick={submit}>{t("confirm")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default Community;
