package main

type ShoppingList struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

type ShoppingListPatch struct {
	Name  *string  `json:"name"`
	Items []string `json:"items"`
}

type ListPushAction struct {
	Items []string `json:"items"`
}
