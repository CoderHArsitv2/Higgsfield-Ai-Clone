import {
  ComposerSkeleton,
  ModelPickerSkeleton,
  CardGridSkeleton,
} from "@/components/ui/Skeleton";

export default function Loading() {
  return (
    <div className="mx-auto flex max-w-[1600px] flex-col gap-6 px-5 py-6 lg:flex-row lg:items-start">
      <aside className="w-full shrink-0 lg:w-[280px]">
        <ModelPickerSkeleton />
      </aside>
      <div className="flex min-w-0 flex-1 flex-col gap-6">
        <ComposerSkeleton />
        <section className="min-w-0">
          <div className="mb-3 border-t border-line pt-5">
            <div className="skeleton h-3 w-36 rounded" />
          </div>
          <CardGridSkeleton count={3} />
        </section>
      </div>
    </div>
  );
}
