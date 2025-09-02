# API Gateways Specification

## Public Gateway
```yaml
routeChecker:
  allowedRoutes:
    - "/api/v1/services"
    - "/api/v1/services/{id}"
    - "/api/v1/services/slug/{slug}"
    - "/api/v1/services/featured"
    - "/api/v1/services/categories"
    - "/api/v1/services/categories/{id}/services"
    - "/api/v1/services/search"
    - "/api/v1/content"
    - "/api/v1/content/{id}"
    - "/api/v1/content/{id}/download"
    - "/api/v1/content/{id}/preview"
    - "/health"
    - "/health/ready"
```

## Admin Gateway
```yaml
routeChecker:
  allowedRoutes:
    - "/admin/api/v1/services"
    - "/admin/api/v1/services/{id}"
    - "/admin/api/v1/services/{id}/publish"
    - "/admin/api/v1/services/{id}/archive"
    - "/admin/api/v1/services/{id}/audit"
    - "/admin/api/v1/services/categories"
    - "/admin/api/v1/services/categories/{id}"
    - "/admin/api/v1/services/categories/{id}/audit"
    - "/admin/api/v1/services/featured-categories"
    - "/admin/api/v1/content"
    - "/admin/api/v1/content/{id}"
    - "/admin/api/v1/content/upload"
    - "/admin/api/v1/content/{id}/reprocess"
    - "/admin/api/v1/content/{id}/status"
    - "/admin/api/v1/content/{id}/audit"
    - "/admin/api/v1/content/processing-queue"
    - "/admin/api/v1/content/analytics"
    - "/health"
    - "/health/ready"
```

