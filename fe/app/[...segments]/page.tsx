import { RoutePage } from "@/components/route-page";

export default async function Page({ params, searchParams }: { params: Promise<{ segments?: string[] }>; searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const [{ segments = [] }, query] = await Promise.all([params, searchParams]);
  const first = (value: string | string[] | undefined) => Array.isArray(value) ? value[0] : value;
  return <RoutePage segments={segments} plannerInitial={{ city: first(query.city), startDate: first(query.startDate), endDate: first(query.endDate), travelType: first(query.travelType) }} />;
}
