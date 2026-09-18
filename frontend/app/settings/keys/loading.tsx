export default function Loading() {
  return (
    <main className="mx-auto max-w-3xl px-5 py-10">
      <div className="skeleton h-10 w-64 rounded" />
      <div className="skeleton mt-4 h-3 w-full max-w-xl rounded" />
      <div className="skeleton mt-2 h-3 w-4/5 max-w-lg rounded" />
      <div className="mt-8 space-y-3">
        {[0, 1, 2, 3].map((i) => (
          <div key={i} className="skeleton h-20 rounded-xl" />
        ))}
      </div>
    </main>
  );
}
