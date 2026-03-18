export default function ProfileLoading() {
  return (
    <div className="max-w-3xl mx-auto px-4 py-8 animate-pulse">
      <p className="text-sm text-gray-500 mb-4">Memuat profil...</p>
      <div className="h-8 w-40 bg-gray-200 rounded mb-6" />
      <div className="rounded-2xl border border-gray-100 bg-white p-6 mb-4">
        <div className="flex items-center gap-4">
          <div className="h-20 w-20 rounded-2xl bg-gray-200" />
          <div className="flex-1 space-y-2">
            <div className="h-5 w-40 bg-gray-200 rounded" />
            <div className="h-4 w-56 bg-gray-200 rounded" />
            <div className="h-4 w-32 bg-gray-200 rounded" />
          </div>
        </div>
      </div>
      <div className="rounded-2xl border border-gray-100 bg-white p-6">
        <div className="h-5 w-48 bg-gray-200 rounded mb-4" />
        <div className="space-y-3">
          <div className="h-4 w-full bg-gray-200 rounded" />
          <div className="h-4 w-5/6 bg-gray-200 rounded" />
        </div>
      </div>
    </div>
  );
}
