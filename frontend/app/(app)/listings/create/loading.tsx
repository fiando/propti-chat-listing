export default function CreateListingLoading() {
  return (
    <div className="max-w-2xl mx-auto px-4 py-8 animate-pulse">
      <p className="text-sm text-gray-500 mb-4">Menyiapkan form iklan...</p>
      <div className="h-8 w-52 bg-gray-200 rounded mb-3" />
      <div className="h-4 w-80 bg-gray-200 rounded mb-6" />

      <div className="flex items-center gap-2 mb-8">
        <div className="h-7 w-7 rounded-full bg-gray-200" />
        <div className="h-1 w-8 bg-gray-200 rounded" />
        <div className="h-7 w-7 rounded-full bg-gray-200" />
        <div className="h-1 w-8 bg-gray-200 rounded" />
        <div className="h-7 w-7 rounded-full bg-gray-200" />
      </div>

      <div className="grid sm:grid-cols-2 gap-4">
        <div className="rounded-2xl border border-gray-100 bg-white p-6 space-y-3">
          <div className="h-10 w-10 rounded-xl bg-gray-200" />
          <div className="h-5 w-32 bg-gray-200 rounded" />
          <div className="h-4 w-full bg-gray-200 rounded" />
          <div className="h-4 w-5/6 bg-gray-200 rounded" />
        </div>
        <div className="rounded-2xl border border-gray-100 bg-white p-6 space-y-3">
          <div className="h-10 w-10 rounded-xl bg-gray-200" />
          <div className="h-5 w-32 bg-gray-200 rounded" />
          <div className="h-4 w-full bg-gray-200 rounded" />
          <div className="h-4 w-5/6 bg-gray-200 rounded" />
        </div>
      </div>
    </div>
  );
}
