export default function OnlineIndicator({ online }: { online: boolean }) {
  return (
    <span className="inline-flex items-center gap-1.5 text-sm">
      <span className={`h-2 w-2 rounded-full ${online ? 'bg-green-500' : 'bg-gray-400'}`} />
      <span className={online ? 'text-green-700' : 'text-gray-500'}>
        {online ? 'Online' : 'Offline'}
      </span>
    </span>
  )
}
