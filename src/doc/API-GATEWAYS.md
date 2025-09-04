# API Gateways Specification

## Public Gateway
```yaml
routeChecker:
  allowedRoutes:
    # Services domain endpoints
    - "/api/v1/services"
    - "/api/v1/services/{id}"
    - "/api/v1/services/slug/{slug}"
    - "/api/v1/services/featured"
    - "/api/v1/services/categories"
    - "/api/v1/services/categories/{id}/services"
    - "/api/v1/services/search"
    # News domain endpoints
    - "/api/v1/news"
    - "/api/v1/news/{id}"
    - "/api/v1/news/slug/{slug}"
    - "/api/v1/news/featured"
    - "/api/v1/news/categories"
    - "/api/v1/news/categories/{id}/news"
    - "/api/v1/news/search"
    # Research domain endpoints
    - "/api/v1/research"
    - "/api/v1/research/{id}"
    - "/api/v1/research/slug/{slug}"
    - "/api/v1/research/featured"
    - "/api/v1/research/categories"
    - "/api/v1/research/categories/{id}/research"
    - "/api/v1/research/search"
    - "/api/v1/research/{id}/report"
    # Events domain endpoints
    - "/api/v1/events"
    - "/api/v1/events/{id}"
    - "/api/v1/events/slug/{slug}"
    - "/api/v1/events/featured"
    - "/api/v1/events/categories"
    - "/api/v1/events/categories/{id}/events"
    - "/api/v1/events/search"
    - "/api/v1/events/{id}/register"
    - "/api/v1/events/{id}/registrations"
    # Inquiries form submission endpoints
    - "/api/v1/inquiries/media"
    - "/api/v1/inquiries/business"
    - "/api/v1/inquiries/donations"
    # Health endpoints
    - "/health"
    - "/health/ready"
```

## Admin Gateway
```yaml
routeChecker:
  allowedRoutes:
    # Services domain admin endpoints
    - "/admin/api/v1/services"
    - "/admin/api/v1/services/{id}"
    - "/admin/api/v1/services/{id}/publish"
    - "/admin/api/v1/services/{id}/archive"
    - "/admin/api/v1/services/{id}/audit"
    - "/admin/api/v1/services/categories"
    - "/admin/api/v1/services/categories/{id}"
    - "/admin/api/v1/services/categories/{id}/audit"
    - "/admin/api/v1/services/featured-categories"
    # News domain admin endpoints
    - "/admin/api/v1/news"
    - "/admin/api/v1/news/{id}"
    - "/admin/api/v1/news/{id}/publish"
    - "/admin/api/v1/news/{id}/archive"
    - "/admin/api/v1/news/{id}/audit"
    - "/admin/api/v1/news/categories"
    - "/admin/api/v1/news/categories/{id}"
    - "/admin/api/v1/news/categories/{id}/audit"
    - "/admin/api/v1/news/featured"
    # Research domain admin endpoints
    - "/admin/api/v1/research"
    - "/admin/api/v1/research/{id}"
    - "/admin/api/v1/research/{id}/publish"
    - "/admin/api/v1/research/{id}/archive"
    - "/admin/api/v1/research/{id}/audit"
    - "/admin/api/v1/research/{id}/report/upload"
    - "/admin/api/v1/research/categories"
    - "/admin/api/v1/research/categories/{id}"
    - "/admin/api/v1/research/categories/{id}/audit"
    - "/admin/api/v1/research/featured"
    # Events domain admin endpoints
    - "/admin/api/v1/events"
    - "/admin/api/v1/events/{id}"
    - "/admin/api/v1/events/{id}/publish"
    - "/admin/api/v1/events/{id}/archive"
    - "/admin/api/v1/events/{id}/audit"
    - "/admin/api/v1/events/categories"
    - "/admin/api/v1/events/categories/{id}"
    - "/admin/api/v1/events/categories/{id}/audit"
    - "/admin/api/v1/events/featured"
    - "/admin/api/v1/events/{id}/registrations"
    - "/admin/api/v1/events/{id}/registrations/{reg_id}"
    # Inquiries admin management endpoints
    - "/admin/api/v1/inquiries/media"
    - "/admin/api/v1/inquiries/media/{id}"
    - "/admin/api/v1/inquiries/media/{id}/audit"
    - "/admin/api/v1/inquiries/business"
    - "/admin/api/v1/inquiries/business/{id}"
    - "/admin/api/v1/inquiries/business/{id}/audit"
    - "/admin/api/v1/inquiries/donations"
    - "/admin/api/v1/inquiries/donations/{id}"
    - "/admin/api/v1/inquiries/donations/{id}/audit"
    # Health endpoints
    - "/health"
    - "/health/ready"
```

