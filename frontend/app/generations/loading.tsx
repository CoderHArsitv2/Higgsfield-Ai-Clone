import { CardGridSkeleton } from "@/components/ui/Skeleton";

export default function Loading() {
  return (
    <main className="mx-auto max-w-[1600px] px-5 py-8">
      <div className="skeleton h-9 w-56 rounded" />
      <div className="skeleton mt-4 h-3 w-96 max-w-full rounded" />
      <div className="mt-9 mb-5 flex gap-1.5">
        {[44, 62, 52, 56, 68].map((w, i) => (
          <div
            key={i}
            className="skeleton h-7 rounded-full"
            style={{ width: w }}
          />
        ))}
      </div>
      <CardGridSkeleton count={6} />
    </main>
  );
}
