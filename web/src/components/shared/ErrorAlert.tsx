export default function ErrorAlert({ message }: { message: string }) {
  return (
    <div className="rounded-md bg-red-50 border border-red-200 p-4">
      <p className="text-sm text-red-800">{message}</p>
    </div>
  )
}
