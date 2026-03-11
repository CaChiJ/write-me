import { NewSessionPage } from "@/features/sessions/NewSessionPage";

export default async function Page({
  searchParams
}: {
  searchParams: Promise<{ mode?: "posting" | "draft" | "assets" }>;
}) {
  const params = await searchParams;
  return <NewSessionPage initialMode={params.mode} />;
}
