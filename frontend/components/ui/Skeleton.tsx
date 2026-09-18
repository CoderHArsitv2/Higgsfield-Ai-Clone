/**
 * Loading placeholders that mirror the real component's geometry.
 *
 * They are sized like what they stand in for -- a 16:9 media box, two lines of
 * caption -- so the page does not jump when the data lands. A spinner in the
 * middle of an empty panel cannot do that.
 */

export function CardSkeleton() {
  return (
    <div className="overflow-hidden rounded-xl border border-line bg-panel">
      <div className="skeleton aspect-video" />
      <div className="space-y-2.5 p-4">
        <div className="skeleton h-3 w-full rounded" />
        <div className="skeleton h-3 w-2/3 rounded" />
        <div className="flex justify-between pt-1">
          <div className="skeleton h-2.5 w-20 rounded" />
          <div className="skeleton h-2.5 w-12 rounded" />
        </div>
      </div>
    </div>
  );
}

export function CardGridSkeleton({ count = 3 }: { count?: number }) {
  return (
    <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
      {Array.from({ length: count }, (_, i) => (
        <CardSkeleton key={i} />
      ))}
    </div>
  );
}

export function ModelPickerSkeleton() {
  return (
    <div className="space-y-4" aria-hidden>
      <div className="skeleton h-9 w-full rounded-lg" />
      <div className="flex gap-1.5">
        {[44, 60, 58, 56].map((w, i) => (
          <div
            key={i}
            className="skeleton h-6 rounded-full"
            style={{ width: w }}
          />
        ))}
      </div>
      {[0, 1].map((g) => (
        <div key={g} className="space-y-1 pt-1">
          <div className="skeleton mb-2 h-3 w-24 rounded" />
          {[0, 1, 2, 3].map((i) => (
            <div key={i} className="space-y-1.5 px-3 py-2">
              <div className="skeleton h-3 w-32 rounded" />
              <div className="skeleton h-2.5 w-44 rounded" />
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}

/** The composer's shape, for the server-rendered route fallback. */
export function ComposerSkeleton() {
  return (
    <div className="rounded-2xl border border-line bg-panel p-5" aria-hidden>
      <div className="skeleton mb-3 h-4 w-40 rounded" />
      <div className="skeleton h-[86px] w-full rounded-xl" />
      <div className="mt-4 grid gap-3 sm:grid-cols-2">
        <div className="skeleton h-14 rounded-lg" />
        <div className="skeleton h-14 rounded-lg" />
      </div>
      <div className="mt-4 flex justify-end">
        <div className="skeleton h-10 w-32 rounded-full" />
      </div>
    </div>
  );
}
