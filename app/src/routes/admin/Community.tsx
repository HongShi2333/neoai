import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card.tsx";
import { useTranslation } from "react-i18next";
import CommunityTable from "@/components/admin/assemblies/CommunityTable.tsx";

function Community() {
  const { t } = useTranslation();

  return (
    <div className={`community`}>
      <Card className={`admin-card community-card`}>
        <CardHeader className={`select-none`}>
          <CardTitle>{t("admin.community.title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <CommunityTable />
        </CardContent>
      </Card>
    </div>
  );
}

export default Community;
