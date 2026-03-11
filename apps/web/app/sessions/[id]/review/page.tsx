import { ReviewPage } from "@/features/sessions/ReviewPage";

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <ReviewPage sessionId={id} />;
}
