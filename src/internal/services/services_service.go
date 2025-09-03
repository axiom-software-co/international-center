package services

type ServicesService struct {
	repository ServicesRepository
}

func NewServicesService(repository ServicesRepository) *ServicesService {
	return &ServicesService{
		repository: repository,
	}
}

func (s *ServicesService) CreateService(title, description, slug, deliveryMode string) (*Service, error) {
	service, err := NewService(title, description, slug, deliveryMode)
	if err != nil {
		return nil, err
	}
	
	err = s.repository.Create(service)
	if err != nil {
		return nil, err
	}
	
	return service, nil
}

func (s *ServicesService) GetService(serviceID string) (*Service, error) {
	return s.repository.GetByID(serviceID)
}

func (s *ServicesService) GetServiceBySlug(slug string) (*Service, error) {
	return s.repository.GetBySlug(slug)
}

func (s *ServicesService) UpdateService(serviceID, title, description, slug, deliveryMode, userID string) (*Service, error) {
	service, err := s.repository.GetByID(serviceID)
	if err != nil {
		return nil, err
	}
	
	service.Title = title
	service.Description = description
	service.Slug = slug
	service.DeliveryMode = DeliveryMode(deliveryMode)
	service.ModifiedBy = userID
	
	if !isValidDeliveryMode(service.DeliveryMode) {
		return nil, err
	}
	
	err = s.repository.Update(service)
	if err != nil {
		return nil, err
	}
	
	return service, nil
}

func (s *ServicesService) PublishService(serviceID, userID string) error {
	service, err := s.repository.GetByID(serviceID)
	if err != nil {
		return err
	}
	
	err = service.Publish(userID)
	if err != nil {
		return err
	}
	
	return s.repository.Update(service)
}

func (s *ServicesService) ArchiveService(serviceID, userID string) error {
	service, err := s.repository.GetByID(serviceID)
	if err != nil {
		return err
	}
	
	err = service.Archive(userID)
	if err != nil {
		return err
	}
	
	return s.repository.Update(service)
}

func (s *ServicesService) AssignServiceCategory(serviceID, categoryID, userID string) error {
	service, err := s.repository.GetByID(serviceID)
	if err != nil {
		return err
	}
	
	err = service.AssignCategory(categoryID, userID)
	if err != nil {
		return err
	}
	
	return s.repository.Update(service)
}

func (s *ServicesService) DeleteService(serviceID, userID string) error {
	return s.repository.Delete(serviceID, userID)
}

func (s *ServicesService) ListServices(offset, limit int) ([]*Service, error) {
	return s.repository.List(offset, limit)
}

func (s *ServicesService) ListServicesByCategory(categoryID string, offset, limit int) ([]*Service, error) {
	return s.repository.ListByCategory(categoryID, offset, limit)
}

func (s *ServicesService) ListPublishedServices(offset, limit int) ([]*Service, error) {
	return s.repository.ListPublished(offset, limit)
}