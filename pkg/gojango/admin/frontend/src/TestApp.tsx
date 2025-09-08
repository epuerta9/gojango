import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

function TestApp() {
  return (
    <div style={{ 
      minHeight: '100vh', 
      backgroundColor: '#f8fafc', 
      padding: '2rem',
      fontFamily: 'system-ui, -apple-system, sans-serif'
    }}>
      <div style={{ maxWidth: '1024px', margin: '0 auto' }}>
        <div style={{
          backgroundColor: 'white',
          borderRadius: '8px',
          padding: '24px',
          boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
          border: '1px solid #e2e8f0'
        }}>
          <h1 style={{ 
            fontSize: '24px', 
            fontWeight: 'bold', 
            marginBottom: '16px',
            color: '#1e293b'
          }}>
            🎉 Gojango Admin Test
          </h1>
          <p style={{ 
            color: '#64748b', 
            marginBottom: '16px' 
          }}>
            React app is loading! Testing with inline styles to check if CSS assets are working.
          </p>
          <button style={{
            backgroundColor: '#3b82f6',
            color: 'white',
            padding: '8px 16px',
            borderRadius: '6px',
            border: 'none',
            fontSize: '14px',
            fontWeight: '500',
            cursor: 'pointer'
          }}>
            Test Button (Inline Styled)
          </button>
          
          <div style={{ marginTop: '24px' }}>
            <h3 style={{ 
              fontSize: '18px', 
              fontWeight: '600', 
              marginBottom: '12px',
              color: '#1e293b'
            }}>
              Now testing Shadcn UI components:
            </h3>
            <Card>
              <CardHeader>
                <CardTitle>Shadcn UI Card</CardTitle>
              </CardHeader>
              <CardContent>
                <p>If this card looks styled, Shadcn UI is working!</p>
                <Button>Shadcn Button</Button>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  )
}

export default TestApp