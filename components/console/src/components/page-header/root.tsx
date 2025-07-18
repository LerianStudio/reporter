import { Collapsible } from '@/components/ui/collapsible'

type RootProps = React.ComponentProps<typeof Collapsible>

export const Root = ({ children, className, ...props }: RootProps) => {
  return (
    <div className="mt-12">
      <Collapsible className={className} {...props}>
        {children}
      </Collapsible>
    </div>
  )
}
