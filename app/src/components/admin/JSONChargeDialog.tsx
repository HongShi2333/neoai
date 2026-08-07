import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button.tsx";
import { Textarea } from "@/components/ui/textarea.tsx";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog.tsx";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select.tsx";
import { toast } from "sonner";
import { Download, Upload, FileJson } from "lucide-react";
import {
  applyChargeJSON,
  exportChargeJSON,
  ChargeJSONMode,
} from "@/admin/api/charge.ts";
import { ChargeProps } from "@/admin/charge.ts";

// JSON charge editor — a new-api style bulk import/export dialog.
//
// Open it from the ChargeWidget toolbar. The textarea pre-fills with
// the current rules as pretty JSON; the admin can edit it and click
// "Apply" to push the entire document back to the server in one shot.
//
// Two modes:
//   replace — wipe existing rules first, then write the new set
//   merge   — keep existing rules, append / overwrite per-model
//
// Both modes go through /admin/charge/json on the backend, which
// validates every row before touching the config file.

type Props = {
  open: boolean;
  onOpenChange: (o: boolean) => void;
  onRefresh: () => void;
};

const sample = `[
  {
    "type": "token",
    "models": ["gpt-4", "gpt-4-turbo", "gpt-4o"],
    "input": 0.03,
    "output": 0.06,
    "anonymous": false
  },
  {
    "type": "times",
    "models": ["dall-e-3"],
    "input": 0,
    "output": 0.04
  },
  {
    "type": "non-billing",
    "models": ["gpt-3.5-turbo-1106"]
  }
]`;

export function JSONChargeDialog({ open, onOpenChange, onRefresh }: Props) {
  const { t } = useTranslation();
  const [text, setText] = useState("");
  const [mode, setMode] = useState<ChargeJSONMode>("replace");
  const [loading, setLoading] = useState(false);

  async function loadCurrent() {
    setLoading(true);
    const resp = await exportChargeJSON();
    setLoading(false);
    if (resp.status && resp.data) {
      // Strip volatile fields so the diff stays clean.
      const cleaned = resp.data.map((r: ChargeProps) => ({
        type: r.type,
        models: r.models,
        input: r.input,
        output: r.output,
        anonymous: r.anonymous,
      }));
      setText(JSON.stringify(cleaned, null, 2));
    } else {
      toast.error(resp.error || "export failed");
    }
  }

  async function apply() {
    let rules: ChargeProps[];
    try {
      rules = JSON.parse(text);
      if (!Array.isArray(rules)) {
        throw new Error("expected a JSON array");
      }
    } catch (e: any) {
      toast.error(`invalid JSON: ${e.message || e}`);
      return;
    }
    setLoading(true);
    const resp = await applyChargeJSON(rules, mode);
    setLoading(false);
    if (resp.status) {
      toast.success(
        t("admin.charge.json-applied", { count: rules.length }) ||
          `applied ${rules.length} rules`,
      );
      onOpenChange(false);
      onRefresh();
    } else {
      toast.error(resp.error || "apply failed");
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle>
            <FileJson className="h-4 w-4 inline mr-1.5" />
            {t("admin.charge.json-editor") || "JSON Pricing Editor"}
          </DialogTitle>
          <DialogDescription>
            {t("admin.charge.json-editor-description") ||
              "Edit pricing rules as a JSON document. Apply in bulk with replace or merge mode."}
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-2 mb-2">
          <Button variant="outline" size="sm" onClick={loadCurrent} disabled={loading}>
            <Download className="h-3.5 w-3.5 mr-1.5" />
            {t("admin.charge.load-current") || "Load current"}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setText(sample)}
          >
            {t("admin.charge.insert-sample") || "Insert sample"}
          </Button>
          <div className="grow" />
          <span className="text-sm text-muted-foreground mr-1">
            {t("admin.charge.mode") || "Mode"}:
          </span>
          <Select
            value={mode}
            onValueChange={(v) => setMode(v as ChargeJSONMode)}
          >
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="replace">
                {t("admin.charge.mode-replace") || "Replace all"}
              </SelectItem>
              <SelectItem value="merge">
                {t("admin.charge.mode-merge") || "Merge"}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>

        <Textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder={sample}
          className="flex-1 min-h-[400px] font-mono text-xs"
          spellCheck={false}
        />

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("cancel")}
          </Button>
          <Button onClick={apply} disabled={loading || !text.trim()}>
            <Upload className="h-4 w-4 mr-1.5" />
            {t("admin.charge.apply") || "Apply"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
