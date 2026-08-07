import { useEffect, useRef, useState } from "react";
import { useSearchParams } from "react-router-dom";
import {
  Channel,
  Message,
  Visibility,
  SendPermission,
  listChannels,
  listMessages,
  postMessage,
  deleteMessage,
  createChannel,
  updateChannel,
  deleteChannel,
  communityWsUrl,
  ChannelForm,
} from "@/api/community.ts";
import { useTranslation } from "react-i18next";
import { useSelector } from "react-redux";
import { selectAdmin, selectAuthenticated, selectUsername } from "@/store/auth.ts";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog.tsx";
import { Button } from "@/components/ui/button.tsx";
import { Input } from "@/components/ui/input.tsx";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select.tsx";
import { Label } from "@/components/ui/label.tsx";
import { Badge } from "@/components/ui/badge.tsx";
import {
  Hash,
  Plus,
  Send,
  Trash2,
  Settings2,
  Lock,
  Globe,
  Users,
  Shield,
} from "lucide-react";
import { format } from "date-fns";
import { toast } from "sonner";

// Community page — a Discord-style channel list + chat pane.
// Authenticated users see channels they have visibility into, and can
// post to channels their send-permission allows. Admins see a "+" button
// to create new channels and a gear icon on each channel to edit / delete.

const VIS_OPTIONS: { value: Visibility; label: string; icon: any }[] = [
  { value: "public", label: "Public (anyone)", icon: Globe },
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
  const auth = useSelector(selectAuthenticated);
  const admin = useSelector(selectAdmin);
  const username = useSelector(selectUsername);
  const [search, setSearch] = useSearchParams();

  const [channels, setChannels] = useState<Channel[]>([]);
  const [currentId, setCurrentId] = useState<number | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [draft, setDraft] = useState("");
  const [editorOpen, setEditorOpen] = useState(false);
  const [editing, setEditing] = useState<Channel | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  // ---- channel list bootstrap ----
  async function refreshChannels() {
    const resp = await listChannels();
    if (resp.status && resp.data) {
      setChannels(resp.data);
      // pick first channel if none selected
      if (resp.data.length > 0 && currentId === null) {
        setCurrentId(resp.data[0].id);
      }
    }
  }

  useEffect(() => {
    if (!auth) return;
    refreshChannels().then();
  }, [auth]);

  // ---- url sync ----
  useEffect(() => {
    const idParam = search.get("c");
    if (idParam) {
      const id = parseInt(idParam, 10);
      if (!isNaN(id)) setCurrentId(id);
    }
  }, [search]);

  useEffect(() => {
    if (currentId !== null) {
      const next = new URLSearchParams(search);
      next.set("c", String(currentId));
      setSearch(next, { replace: true });
    }
  }, [currentId]);

  // ---- messages bootstrap + websocket ----
  useEffect(() => {
    if (currentId === null) return;
    listMessages(currentId).then((resp) => {
      if (resp.status) setMessages(resp.data || []);
    });
    setMessages([]);

    // open websocket
    const ws = new WebSocket(communityWsUrl());
    wsRef.current = ws;
    ws.onopen = () => {
      ws.send(
        JSON.stringify({ action: "subscribe", channel_id: currentId }),
      );
    };
    ws.onmessage = (ev) => {
      try {
        const payload = JSON.parse(ev.data);
        if (payload.type === "message" && payload.data) {
          setMessages((prev) =>
            prev.some((m) => m.id === payload.data.id)
              ? prev
              : [...prev, payload.data as Message],
          );
        } else if (payload.type === "delete" && payload.message_id) {
          setMessages((prev) =>
            prev.filter((m) => m.id !== payload.message_id),
          );
        } else if (payload.type === "error") {
          toast.error(payload.message || "websocket error");
        }
      } catch {
        // ignore malformed frames
      }
    };
    return () => {
      ws.close();
    };
  }, [currentId]);

  // ---- autoscroll ----
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // ---- actions ----
  async function send() {
    if (currentId === null || draft.trim() === "") return;
    const text = draft.trim();
    setDraft("");
    const resp = await postMessage(currentId, text);
    if (!resp.status) {
      toast.error(resp.error || "failed to send");
      setDraft(text);
    }
  }

  async function removeMessage(id: number) {
    const resp = await deleteMessage(id);
    if (!resp.status) {
      toast.error(resp.error || "failed to delete");
    }
  }

  function openEditor(ch: Channel | null) {
    setEditing(ch);
    setEditorOpen(true);
  }

  async function submitEditor(form: ChannelForm) {
    if (editing) {
      const resp = await updateChannel(editing.id, form);
      if (resp.status) {
        toast.success("channel updated");
        setEditorOpen(false);
        refreshChannels();
      } else {
        toast.error(resp.error || "update failed");
      }
    } else {
      const resp = await createChannel(form);
      if (resp.status) {
        toast.success("channel created");
        setEditorOpen(false);
        refreshChannels();
      } else {
        toast.error(resp.error || "create failed");
      }
    }
  }

  async function removeChannel(id: number) {
    if (!confirm(t("community.confirm-delete")) || !admin) return;
    const resp = await deleteChannel(id);
    if (resp.status) {
      toast.success("channel deleted");
      refreshChannels();
      if (currentId === id) setCurrentId(null);
    } else {
      toast.error(resp.error || "delete failed");
    }
  }

  if (!auth) {
    return (
      <div className="flex flex-col items-center justify-center h-full p-8 text-center">
        <Users className="w-12 h-12 text-muted-foreground mb-4" />
        <h2 className="text-xl font-semibold mb-2">
          {t("community.login-required")}
        </h2>
        <p className="text-muted-foreground">
          {t("community.login-required-description")}
        </p>
      </div>
    );
  }

  const current = channels.find((c) => c.id === currentId);

  return (
    <div className="flex h-full w-full">
      {/* Sidebar */}
      <div className="w-64 border-r flex flex-col bg-muted/20">
        <div className="px-4 py-3 border-b flex items-center justify-between">
          <span className="font-semibold">{t("community.channels")}</span>
          {admin && (
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => openEditor(null)}
              title={t("community.create")}
            >
              <Plus className="h-4 w-4" />
            </Button>
          )}
        </div>
        <div className="flex-1 overflow-y-auto">
          {channels.length === 0 && (
            <div className="px-4 py-6 text-sm text-muted-foreground">
              {t("community.no-channels")}
            </div>
          )}
          {channels.map((ch) => (
            <button
              key={ch.id}
              onClick={() => setCurrentId(ch.id)}
              className={`w-full text-left px-4 py-2 hover:bg-muted/40 flex items-center gap-2 ${
                ch.id === currentId ? "bg-muted/60" : ""
              }`}
            >
              <Hash className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="truncate flex-1">{ch.name}</span>
              {admin && (
                <span
                  role="button"
                  onClick={(e) => {
                    e.stopPropagation();
                    openEditor(ch);
                  }}
                  className="opacity-0 group-hover:opacity-100 hover:opacity-100"
                >
                  <Settings2 className="h-3.5 w-3.5 text-muted-foreground" />
                </span>
              )}
              <VisibilityBadge v={ch.visibility} />
            </button>
          ))}
        </div>
      </div>

      {/* Chat pane */}
      <div className="flex-1 flex flex-col">
        {current ? (
          <>
            <div className="px-4 py-3 border-b flex items-center gap-2">
              <Hash className="h-4 w-4 text-muted-foreground" />
              <span className="font-semibold">{current.name}</span>
              {current.topic && (
                <span className="text-sm text-muted-foreground truncate">
                  — {current.topic}
                </span>
              )}
              {admin && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="ml-auto"
                  onClick={() => openEditor(current)}
                >
                  <Settings2 className="h-3.5 w-3.5 mr-1" />
                  {t("community.edit")}
                </Button>
              )}
            </div>
            <div className="flex-1 overflow-y-auto px-4 py-3 space-y-2">
              {messages.map((m) => (
                <div
                  key={m.id}
                  className="flex gap-3 group hover:bg-muted/20 rounded px-2 py-1"
                >
                  <div className="w-9 h-9 rounded-full bg-muted flex items-center justify-center text-sm font-semibold shrink-0">
                    {m.username.charAt(0).toUpperCase()}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-baseline gap-2">
                      <span className="font-semibold text-sm">
                        {m.username}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {format(new Date(m.created_at), "MM-dd HH:mm")}
                      </span>
                      {m.edited_at && (
                        <span className="text-xs text-muted-foreground italic">
                          (edited)
                        </span>
                      )}
                      {(m.username === username || admin) && (
                        <button
                          onClick={() => removeMessage(m.id)}
                          className="opacity-0 group-hover:opacity-100 ml-auto text-destructive hover:text-destructive/80"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </button>
                      )}
                    </div>
                    <div className="text-sm whitespace-pre-wrap break-words">
                      {m.content}
                    </div>
                  </div>
                </div>
              ))}
              <div ref={messagesEndRef} />
            </div>
            <div className="px-4 py-3 border-t flex items-center gap-2">
              <Input
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && !e.shiftKey) {
                    e.preventDefault();
                    send();
                  }
                }}
                placeholder={t("community.message-placeholder")}
                className="flex-1"
              />
              <Button onClick={send} size="icon">
                <Send className="h-4 w-4" />
              </Button>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-muted-foreground">
            {t("community.select-channel")}
          </div>
        )}
      </div>

      {/* Editor dialog */}
      <ChannelEditorDialog
        open={editorOpen}
        onOpenChange={setEditorOpen}
        channel={editing}
        onSubmit={submitEditor}
        onDelete={removeChannel}
        admin={admin}
      />
    </div>
  );
}

function VisibilityBadge({ v }: { v: Visibility }) {
  const opt = VIS_OPTIONS.find((o) => o.value === v);
  if (!opt) return null;
  const Icon = opt.icon;
  return (
    <Badge variant="outline" className="text-[10px] px-1 py-0">
      <Icon className="h-2.5 w-2.5 mr-0.5" />
      {opt.label.split(" ")[0]}
    </Badge>
  );
}

function ChannelEditorDialog(props: {
  open: boolean;
  onOpenChange: (o: boolean) => void;
  channel: Channel | null;
  onSubmit: (form: ChannelForm) => void;
  onDelete: (id: number) => void;
  admin: boolean;
}) {
  const { open, onOpenChange, channel, onSubmit, onDelete, admin } = props;
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
      toast.error("name required");
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
              ? t("community.edit-channel")
              : t("community.create-channel")}
          </DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          <div>
            <Label>{t("community.name")}</Label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="general"
            />
          </div>
          <div>
            <Label>{t("community.topic")}</Label>
            <Input
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              placeholder={t("community.topic-placeholder")}
            />
          </div>
          <div>
            <Label>{t("community.visibility")}</Label>
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
              <Label>{t("community.visible-roles")}</Label>
              <Input
                value={visibleRoles}
                onChange={(e) => setVisibleRoles(e.target.value)}
                placeholder="normal,basic,standard,pro,admin"
              />
              <p className="text-xs text-muted-foreground mt-1">
                {t("community.visible-roles-tip")}
              </p>
            </div>
          )}
          {visibility === "whitelist" && (
            <div>
              <Label>{t("community.visible-users")}</Label>
              <Input
                value={visibleUsers}
                onChange={(e) => setVisibleUsers(e.target.value)}
                placeholder="1,2,3"
              />
              <p className="text-xs text-muted-foreground mt-1">
                {t("community.user-ids-tip")}
              </p>
            </div>
          )}
          <div>
            <Label>{t("community.send-permission")}</Label>
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
              <Label>{t("community.senders")}</Label>
              <Input
                value={senders}
                onChange={(e) => setSenders(e.target.value)}
                placeholder="1,2,3"
              />
              <p className="text-xs text-muted-foreground mt-1">
                {t("community.user-ids-tip")}
              </p>
            </div>
          )}
        </div>
        <DialogFooter>
          {channel && admin && (
            <Button
              variant="destructive"
              className="mr-auto"
              onClick={() => onDelete(channel.id)}
            >
              <Trash2 className="h-4 w-4 mr-1" />
              {t("community.delete")}
            </Button>
          )}
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
