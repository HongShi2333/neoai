import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card.tsx";
import { Button } from "@/components/ui/button.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Label } from "@/components/ui/label.tsx";
import { Badge } from "@/components/ui/badge.tsx";
import { Loader2, Save, Globe, RefreshCw } from "lucide-react";
import { toast } from "sonner";
import axios from "axios";
import { getErrorMessage } from "@/utils/base.ts";

// Admin OAuth configuration page.
//
// Visual style matches the other admin pages: single `<Card>` with a
// plain `CardHeader` + `CardTitle`, and `CardContent` containing a
// toolbar (refresh + save buttons) above the form.

type ProviderForm = {
  client_id: string;
  client_secret: string;
};

type OAuthConfig = {
  linuxdo: ProviderForm;
  github: ProviderForm;
  frontend_url: string;
};

const defaultConfig: OAuthConfig = {
  linuxdo: { client_id: "", client_secret: "" },
  github: { client_id: "", client_secret: "" },
  frontend_url: "",
};

function OAuthPage() {
  const { t } = useTranslation();
  const [config, setConfig] = useState<OAuthConfig>(defaultConfig);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  async function load() {
    setLoading(true);
    try {
      const r = await axios.get("/admin/oauth/config");
      const data = r.data || {};
      setConfig({
        linuxdo: data.linuxdo || { client_id: "", client_secret: "" },
        github: data.github || { client_id: "", client_secret: "" },
        frontend_url: data.frontend_url || "",
      });
    } catch (e) {
      toast.error(getErrorMessage(e));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function save() {
    setSaving(true);
    try {
      await axios.post("/admin/oauth/config", config);
      toast.success(t("admin.oauth.saved"));
    } catch (e) {
      toast.error(getErrorMessage(e));
    } finally {
      setSaving(false);
    }
  }

  function update<K extends keyof OAuthConfig>(
    key: K,
    value: OAuthConfig[K],
  ) {
    setConfig((prev) => ({ ...prev, [key]: value }));
  }

  function updateProvider(
    provider: "linuxdo" | "github",
    field: keyof ProviderForm,
    value: string,
  ) {
    setConfig((prev) => ({
      ...prev,
      [provider]: { ...prev[provider], [field]: value },
    }));
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className={`oauth`}>
      <Card className={`admin-card oauth-card`}>
        <CardHeader className={`select-none`}>
          <CardTitle>{t("admin.oauth.title")}</CardTitle>
          <CardDescription>{t("admin.oauth.description")}</CardDescription>
        </CardHeader>
        <CardContent>
          {/* Toolbar */}
          <div className="flex items-center justify-end mb-4">
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={load}
                disabled={loading}
              >
                <RefreshCw className="h-4 w-4 mr-1.5" />
                {t("refresh") || "Refresh"}
              </Button>
              <Button size="sm" onClick={save} disabled={saving}>
                {saving ? (
                  <Loader2 className="h-4 w-4 animate-spin mr-1.5" />
                ) : (
                  <Save className="h-4 w-4 mr-1.5" />
                )}
                {t("save") || "Save"}
              </Button>
            </div>
          </div>

          <div className="space-y-6">
            {/* LinuxDO */}
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <span className="font-bold text-[#f5a623] text-lg">L</span>
                <h3 className="font-medium">LinuxDO</h3>
                <Badge variant={config.linuxdo.client_id ? "default" : "outline"}>
                  {config.linuxdo.client_id
                    ? t("admin.oauth.enabled")
                    : t("admin.oauth.disabled")}
                </Badge>
              </div>
              <div>
                <Label>Client ID</Label>
                <Input
                  value={config.linuxdo.client_id}
                  onChange={(e) =>
                    updateProvider("linuxdo", "client_id", e.target.value)
                  }
                  placeholder={t("admin.oauth.client-id-placeholder") || ""}
                />
              </div>
              <div>
                <Label>Client Secret</Label>
                <Input
                  type="password"
                  value={config.linuxdo.client_secret}
                  onChange={(e) =>
                    updateProvider("linuxdo", "client_secret", e.target.value)
                  }
                  placeholder={t("admin.oauth.client-secret-placeholder") || ""}
                />
              </div>
              <p className="text-xs text-muted-foreground">
                {t("admin.oauth.redirect-hint")}{" "}
                <code className="bg-muted px-1.5 py-0.5 rounded">
                  {config.frontend_url || "<backend_url>"}
                  /api/oauth/linuxdo/callback
                </code>
              </p>
            </div>

            <hr />

            {/* GitHub */}
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4" />
                <h3 className="font-medium">GitHub</h3>
                <Badge variant={config.github.client_id ? "default" : "outline"}>
                  {config.github.client_id
                    ? t("admin.oauth.enabled")
                    : t("admin.oauth.disabled")}
                </Badge>
              </div>
              <div>
                <Label>Client ID</Label>
                <Input
                  value={config.github.client_id}
                  onChange={(e) =>
                    updateProvider("github", "client_id", e.target.value)
                  }
                  placeholder={t("admin.oauth.client-id-placeholder") || ""}
                />
              </div>
              <div>
                <Label>Client Secret</Label>
                <Input
                  type="password"
                  value={config.github.client_secret}
                  onChange={(e) =>
                    updateProvider("github", "client_secret", e.target.value)
                  }
                  placeholder={t("admin.oauth.client-secret-placeholder") || ""}
                />
              </div>
              <p className="text-xs text-muted-foreground">
                {t("admin.oauth.redirect-hint")}{" "}
                <code className="bg-muted px-1.5 py-0.5 rounded">
                  {config.frontend_url || "<backend_url>"}
                  /api/oauth/github/callback
                </code>
              </p>
            </div>

            <hr />

            {/* Frontend redirect URL */}
            <div className="space-y-2">
              <Label>{t("admin.oauth.frontend-url")}</Label>
              <Input
                value={config.frontend_url}
                onChange={(e) => update("frontend_url", e.target.value)}
                placeholder="https://app.example.com"
              />
              <p className="text-xs text-muted-foreground">
                {t("admin.oauth.frontend-url-tip")}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default OAuthPage;
