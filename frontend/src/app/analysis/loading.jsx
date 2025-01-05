export default function Loading() {
  return (
    <div className="min-h-screen bg-gray-900 flex flex-col items-center justify-center">
      <div className="text-green-400 text-2xl font-bold mb-4">
        Analyzing Resume
      </div>
      <div className="w-16 h-16 border-4 border-green-400 border-t-transparent rounded-full animate-spin" />
    </div>
  );
}
