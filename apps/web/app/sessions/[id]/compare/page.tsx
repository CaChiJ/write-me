import { ComparePage } from "@/features/sessions/ComparePage";

export default async function Page({
  params,
  searchParams
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ questionId?: string }>;
}) {
  const { id } = await params;
  const query = await searchParams;
  return <ComparePage sessionId={id} initialQuestionId={query.questionId ?? ""} />;
}
