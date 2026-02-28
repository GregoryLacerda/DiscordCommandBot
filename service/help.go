package service

func (s *Service) GetHelpMessage() (string, error) {

	response := `📚 **Comandos Disponíveis:**

				**!ping** - Testa se o bot está respondendo
				**!hello** / **!oi** - Recebe uma saudação
				**!calc <expressão>** - Calcula uma expressão matemática (ex: !calc 2 + 2)
				**!info** - Mostra informações sobre você e o sistema
				**!help** - Mostra esta mensagem de ajuda`
	return response, nil
}
