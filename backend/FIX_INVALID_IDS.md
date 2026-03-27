# Correção de IDs de Produtos Inválidos

## Problema
O sistema estava exibindo erros como "invalid product ID format" ou "invalid uuid leng: 9" quando tentava baixar produtos da biblioteca do usuário. Isso ocorre porque há IDs de produtos inválidos (não UUIDs válidos) no banco de dados.

## Soluções Implementadas

### 1. Validação no Handler (Já implementada)
- Adicionada validação de UUID no início do endpoint `/library/{id}/download`
- Retorna mensagem de erro clara ao invés de erro confuso
- Preveniu que IDs inválidos alcancem o serviço

### 2. Validação no Service (Já implementada)
- Adicionada mensagem de erro mais descritiva nos logs
- Melhoria na identificação da causa raiz do erro

### 3. Scripts de Correção de Dados

São fornecidos dois scripts SQL para ajudar a identificar e corrigir os dados inválidos:

1. `fix_invalid_product_ids.sql` - Script para identificar registros problemáticos
2. `cleanup_invalid_ids.sql` - Script para remover registros com IDs inválidos

## Passos para Corrigir os Dados

### Opção 1: Limpeza Manual (Recomendada)
1. Execute o script de identificação para ver quais registros estão inválidos:
   ```bash
   psql -d pixelcraft -f database/fix_invalid_product_ids.sql
   ```

2. Revise os dados manualmente para entender a causa raiz

3. Execute o script de limpeza se a remoção dos dados for aceitável:
   ```bash
   psql -d pixelcraft -f database/cleanup_invalid_ids.sql
   ```

### Opção 2: Execução Direta
Execute diretamente o script de limpeza (isso removerá permanentemente registros inválidos):
```bash
psql -d pixelcraft -f database/cleanup_invalid_ids.sql
```

## Prevenção Futura
- O sistema agora valida corretamente os UUIDs antes de processar
- Mensagens de erro claras ajudam na identificação de problemas
- Considerar adicionar constraints no banco de dados para impedir inserção de UUIDs inválidos no futuro

## Verificação
Após aplicar as correções:
1. Reinicie o servidor backend
2. Teste o download de produtos na página de downloads
3. Verifique se os produtos aparecem corretamente na biblioteca do usuário