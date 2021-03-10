
// DeleteNamespaceRequest is the input object for NamespaceService.DeleteNamespace
interface DeleteNamespaceRequest {
  
  
    
    
    ID : string;
  
}

// DeleteNamespaceRequest is the output object for NamespaceService.DeleteNamespace
interface DeleteNamespaceResponse {
  
  
    
    
    Error : string;
  
}

// GetNamespaceRequest is the input object for GetNamespaceService.GetNamespace
interface GetNamespaceRequest {
  
  
    
    
    ID : string;
  
}

// GetNamespaceResponse is the output object for GetNamespaceService.GetNamespace
interface GetNamespaceResponse {
  
  
    
    
    
    Namespaces : any[];
  
    
    
    Error : string;
  
}

// PutNamespaceRequest is the input object for NamespaceService.PutNamespace
interface PutNamespaceRequest {
  
  
    
    
    
    Namespace : any;
  
}

// PutNamespaceResponse is the output object for NamespaceService.PutNamespace
interface PutNamespaceResponse {
  
  
    
    
    Error : string;
  
}

// SearchNamespaceRequest is the input object for NamespaceService.Search
interface SearchNamespaceRequest {
  
  
    
    
    Prefix : string;
  
    
    
    BaseURI : string;
  
}

// SearchNamespaceResponse is the output object for NamespaceService.Search
interface SearchNamespaceResponse {
  
  
    
    
    
    Hits : any[];
  
    
    
    More : boolean;
  
    
    
    Error : string;
  
}



// NamespaceService allows you to programmatically manage namespaces
interface NamespaceService {
	
	// DeletetNamespace deletes a Namespace
  deleteNamespace(deleteNamespaceRequest: Partial<DeleteNamespaceRequest>) : Promise<DeleteNamespaceResponse>
  
	// GetNamespace gets a Namespace
  getNamespace(getNamespaceRequest: Partial<GetNamespaceRequest>) : Promise<GetNamespaceResponse>
  
	// PutNamespace stores a Namespace
  putNamespace(putNamespaceRequest: Partial<PutNamespaceRequest>) : Promise<PutNamespaceResponse>
  
	// Search returns a filtered list of Namespaces
  search(searchNamespaceRequest: Partial<SearchNamespaceRequest>) : Promise<SearchNamespaceResponse>
  
}


