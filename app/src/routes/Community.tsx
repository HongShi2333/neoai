import "@/assets/pages/community.less";
import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useSelector } from "react-redux";
import { selectAuthenticated, selectUsername } from "@/store/auth.ts";
import router from "@/router.tsx";
import {
  CommunityChannel,
  CommunityMessage,
  listChannels,
  listMessages,
  openCommunitySocket,
  sendMessage,
} from "@/api/community.ts";
import { Button } from "@/components/ui/button.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Badge } from "@/components/ui/badge.tsx";
import { ScrollArea } from "@/components/ui/scroll-area.tsx";
import { cn } from "@/components/ui/lib/utils.ts";
import Avatar from "@/components/Avatar.tsx";
import Markdown from "@/components/Markdown.tsx";
import {
  Hash,
  Lock,
  Globe,
  Send,
  RotateCw,
  MessageCircle,
} from "lucide-react";
import { useEffectAsync } from "@/utils/hook.ts";
import { toast } from "sonner";

function formatTime(raw: string): string {
  if (!raw) return "";
  const d = new Date(raw);
  if (isNaN(d.getTime())) return raw;
  const now = new Date();
  const sameDay =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate();
  const time = `${String(d.getHours()).padStart(2, "0")}:${String(
    d.getMinutes(),
  ).padStart(2, "0")}`;
  if (sameDay) return time;
  return `${d.getMonth() + 1}/${d.getDate()} ${time}`;
}

type ChannelListProps = {
  channels: CommunityChannel[];
  activeId: number | null;
  onSelect: (ch: CommunityChannel) => void;
  loading: boolean;
  onRefresh: () => void;
};

function ChannelList({
  channels,
  activeId,
  onSelect,
  loading,
  onRefresh,
}: ChannelListProps) {
  const { t } = useTranslation();

  return (
    <div className={`community-sidebar`}>
      <div className={`community-sidebar-header`}>
        <span className={`community-sidebar-title`}>
          {t("community.title")}
        </span>
        <Button
          variant={`ghost`}
          size={`icon`}
          className={`h-7 w-7`}
          onClick={onRefresh}
        >
          <RotateCw className={cn(`h-3.5 w-3.5`, loading && `animate-spin`)} />
        </Button>
      </div>
      <ScrollArea className={`community-channel-list`}>
        {channels.length === 0 ? (
          <p className={`community-empty`}>{t("community.empty")}</p>
        ) : (
          channels.map((ch) => (
            <div
              key={ch.id}
              className={cn(
                `community-channel-item`,
                activeId === ch.id && `active`,
              )}
              onClick={() => onSelect(ch)}
            >
              <div className={`community-channel-icon`}>
                {ch.visibility === "private" ? (
                  <Lock className={`h-3.5 w-3.5`} />
                ) : (
                  <Globe className={`h-3.5 w-3.5`} />
                )}
              </div>
              <div className={`community-channel-meta`}>
                <div className={`community-channel-name`}>
                  <Hash className={`h-3 w-3 inline mr-0.5`} />
                  {ch.name}
                </div>
                {ch.topic && (
                  <div className={`community-channel-topic`}>{ch.topic}</div>
                )}
              </div>
            </div>
          ))
        )}
      </ScrollArea>
    </div>
  );
}

type MessageItemProps = {
  msg: CommunityMessage;
  self: string;
};

function MessageItem({ msg, self }: MessageItemProps) {
  const isSelf = msg.sender_username === self;
  return (
    <div className={cn(`community-message`, isSelf && `self`)}>
      <Avatar
        username={msg.sender_username}
        className={`community-message-avatar`}
      />
      <div className={`community-message-body`}>
        <div className={`community-message-header`}>
          <span className={`community-message-sender`}>
            {msg.sender_username}
          </span>
          <span className={`community-message-time`}>
            {formatTime(msg.created_at)}
          </span>
        </div>
        <div className={`community-message-content`}>
          <Markdown>{msg.content}</Markdown>
        </div>
      </div>
    </div>
  );
}

type ChatPanelProps = {
  channel: CommunityChannel | null;
  messages: CommunityMessage[];
  self: string;
  canPost: boolean;
  onSend: (content: string) => void;
  loadingMessages: boolean;
};

function ChatPanel({
  channel,
  messages,
  self,
  canPost,
  onSend,
  loadingMessages,
}: ChatPanelProps) {
  const { t } = useTranslation();
  const [draft, setDraft] = useState("");
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = scrollRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [messages, channel?.id]);

  if (!channel) {
    return (
      <div className={`community-chat-panel empty`}>
        <div className={`community-placeholder`}>
          <MessageCircle className={`h-10 w-10 mb-2`} />
          <p>{t("community.select-channel")}</p>
        </div>
      </div>
    );
  }

  const submit = () => {
    const content = draft.trim();
    if (!content) return;
    onSend(content);
    setDraft("");
  };

  return (
    <div className={`community-chat-panel`}>
      <div className={`community-chat-header`}>
        <div className={`community-chat-title`}>
          <Hash className={`h-4 w-4`} />
          <span>{channel.name}</span>
          {channel.visibility === "private" ? (
            <Badge variant={`secondary`} className={`ml-2`}>
              <Lock className={`h-3 w-3 mr-1`} />
              {t("community.private")}
            </Badge>
          ) : (
            <Badge variant={`outline`} className={`ml-2`}>
              <Globe className={`h-3 w-3 mr-1`} />
              {t("community.public")}
            </Badge>
          )}
        </div>
        {channel.topic && (
          <div className={`community-chat-topic`}>{channel.topic}</div>
        )}
      </div>
      <div className={`community-chat-body`} ref={scrollRef}>
        {loadingMessages ? (
          <div className={`community-placeholder`}>
            <RotateCw className={`h-5 w-5 animate-spin`} />
          </div>
        ) : messages.length === 0 ? (
          <div className={`community-placeholder`}>
            <p>{t("community.no-messages")}</p>
          </div>
        ) : (
          messages.map((m) => (
            <MessageItem key={m.id} msg={m} self={self} />
          ))
        )}
      </div>
      <div className={`community-chat-footer`}>
        <Input
          value={draft}
          placeholder={
            canPost
              ? t("community.input-placeholder")
              : t("community.readonly")
          }
          disabled={!canPost}
          onChange={(e) => setDraft(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              submit();
            }
          }}
        />
        <Button onClick={submit} disabled={!canPost || !draft.trim()}>
          <Send className={`h-4 w-4 mr-1`} />
          {t("community.send")}
        </Button>
      </div>
    </div>
  );
}

function Community() {
  const { t } = useTranslation();
  const auth = useSelector(selectAuthenticated);
  const self = useSelector(selectUsername) || "";

  const [channels, setChannels] = useState<CommunityChannel[]>([]);
  const [active, setActive] = useState<CommunityChannel | null>(null);
  const [messages, setMessages] = useState<CommunityMessage[]>([]);
  const [loadingChannels, setLoadingChannels] = useState(false);
  const [loadingMessages, setLoadingMessages] = useState(false);

  const socketRef = useRef<ReturnType<typeof openCommunitySocket> | null>(null);

  useEffect(() => {
    if (!auth) {
      router.navigate("/login");
    }
  }, [auth]);

  const refreshChannels = async () => {
    setLoadingChannels(true);
    const resp = await listChannels();
    setLoadingChannels(false);
    if (resp.status) {
      const list = resp.data || [];
      setChannels(list);
      if (list.length > 0 && !active) {
        setActive(list[0]);
      }
    } else if (resp.error) {
      toast.error(resp.error);
    }
  };

  useEffectAsync(async () => {
    if (auth) await refreshChannels();
  }, [auth]);

  // load messages when active channel changes
  useEffectAsync(async () => {
    if (!active) {
      setMessages([]);
      return;
    }
    setLoadingMessages(true);
    const resp = await listMessages(active.id, 0, 50);
    setLoadingMessages(false);
    if (resp.status) {
      setMessages(resp.data || []);
    } else {
      setMessages([]);
    }
  }, [active?.id]);

  // open websocket for the active channel
  useEffect(() => {
    if (socketRef.current) {
      socketRef.current.close();
      socketRef.current = null;
    }
    if (!active) return;

    const socket = openCommunitySocket(
      active.id,
      (msg) => {
        setMessages((prev) => {
          if (prev.some((m) => m.id === msg.id)) return prev;
          return [...prev, msg];
        });
      },
      () => {
        /* closed */
      },
    );
    socketRef.current = socket;

    return () => {
      socket.close();
      socketRef.current = null;
    };
  }, [active?.id]);

  const canPost = useMemo(() => {
    if (!active) return false;
    // backend enforces the actual permission; the UI hint is best-effort.
    return true;
  }, [active]);

  const onSend = async (raw: string) => {
    if (!active) return;
    const resp = await sendMessage(active.id, raw);
    if (!resp.status) {
      toast.error(resp.error || t("community.send-failed"));
      return;
    }
    if (resp.message) {
      setMessages((prev) => {
        if (prev.some((m) => m.id === resp.message!.id)) return prev;
        return [...prev, resp.message!];
      });
    }
  };

  return (
    <div className={`community-page flex flex-row flex-1`}>
      <ChannelList
        channels={channels}
        activeId={active?.id ?? null}
        onSelect={setActive}
        loading={loadingChannels}
        onRefresh={refreshChannels}
      />
      <ChatPanel
        channel={active}
        messages={messages}
        self={self}
        canPost={canPost}
        onSend={onSend}
        loadingMessages={loadingMessages}
      />
    </div>
  );
}

export default Community;
