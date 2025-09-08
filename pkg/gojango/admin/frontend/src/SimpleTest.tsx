function SimpleTest() {
  return (
    <div style={{ padding: '20px', fontFamily: 'system-ui' }}>
      <h1>🎉 React is Working!</h1>
      <p>If you can see this, React is loading correctly.</p>
      <button onClick={() => alert('Button clicked!')}>
        Test Button
      </button>
    </div>
  )
}

export default SimpleTest