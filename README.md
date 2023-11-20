
## Distribution

### Binary

1. Run `make distribute`
2. Upload to aws
3. Invalidate cache on CloudFront

### Add-on

1. Bump version in docker-addon/Dockerfile under labels
2. Push code, tag & push tags
3. Watch github runners to complete
4. Test add-on locally
5. Update add-on repository with new version