# Gojango Examples

This directory contains example projects and tests to validate Gojango functionality.

## End-to-End Test

The `test-e2e.sh` script validates the complete Gojango workflow:

1. **Project Creation** - Tests `gojango new` command
2. **App Generation** - Tests `gojango startapp` command  
3. **Project Structure** - Validates all files are created correctly
4. **Compilation** - Ensures generated projects compile successfully
5. **Runtime** - Verifies the application starts and responds correctly
6. **CLI Commands** - Tests all management commands work

### Running the Test

```bash
# From the project root
./examples/test-e2e.sh

# Or with custom parameters
./examples/test-e2e.sh --project-name myblog --frontend react --database mysql
```

## Example Projects

### Blog Example (Phase 2+)
A complete blog application demonstrating:
- Project structure
- App-based architecture
- Model definitions
- Views and templates
- Static files
- Admin interface

### API Example (Phase 8+)
REST and gRPC API example showing:
- API endpoints
- Auto-generated clients
- Documentation
- Authentication
- Testing

### E-commerce Example (Phase 10+)
Full e-commerce application featuring:
- Multiple apps (products, orders, payments)
- Background tasks
- Signals and events
- Advanced admin
- Frontend integration

## Testing Matrix

The examples test various combinations:

| Project | Frontend | Database | Features | Status |
|---------|----------|----------|----------|---------|
| Simple  | htmx     | sqlite   | basic    | ✅ Phase 1 |
| Blog    | htmx     | postgres | admin    | 🚧 Phase 2 |
| API     | none     | postgres | api      | 🚧 Phase 8 |
| SaaS    | react    | postgres | all      | 🚧 Phase 15 |

## Contributing

When adding new examples:

1. Create a descriptive directory name
2. Include a README.md with setup instructions
3. Add the example to the test matrix
4. Ensure it works with the current phase
5. Document any special requirements