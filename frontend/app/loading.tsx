export default function Loading() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-[#F8F9FA]">
      <div className="flex flex-col items-center gap-4">
        <div className="relative">
          <div className="w-14 h-14 rounded-full border-4 border-brand-light border-t-brand-primary animate-spin" />
        </div>
        <p className="text-brand-secondary font-medium text-sm">Memuat...</p>
      </div>
    </div>
  );
}
