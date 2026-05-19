# Impressao em PDF A4

Esta pasta contem versoes preparadas para impressao:

- arquivos `.md` individuais com estilo A4 embutido
- um arquivo consolidado `GO-TWEETS-A4-COMPLETE.md`
- uma apostila consolidada `GO-TWEETS-APOSTILA-PROFISSIONAL-A4.md`
- um resumo de revisao rapida `GO-TWEETS-SUMARIO-EXECUTIVO-A4.md`
- addendum com a atualizacao da arquitetura e da estrategia de testes na apostila profissional

## Melhor opcao para gerar o PDF

Para uma edicao mais simples, use `GO-TWEETS-A4-COMPLETE.md`.

Para uma edicao com visual de apostila, use `GO-TWEETS-APOSTILA-PROFISSIONAL-A4.md`.

Para revisao curta e retomada de estudo, use `GO-TWEETS-SUMARIO-EXECUTIVO-A4.md`.

## O que a versao profissional inclui

- capa editorial
- folha de rosto
- pagina de uso
- ficha tecnica
- sumario consolidado
- abertura de capitulo
- quebra de pagina entre capitulos
- layout mais proximo de apostila impressa
- atualizacao da arquitetura atual do projeto
- secao dedicada a testes unitarios e de integracao
- aprofundamento maximo dos testes com leitura guiada arquivo por arquivo
- capitulo final consolidando as boas praticas aprendidas
- apendice de consulta rapida com palavras-chave do Go e quando usar

## O que os arquivos consolidados entregam

- capa
- sumario
- quebra de pagina entre capitulos
- layout de papel A4

## Exportacao para PDF dentro do projeto

Foi adicionado o script `export-apostila-pdf.ps1` nesta pasta.

Exemplo de uso no PowerShell:

```powershell
cd c:\Users\marco\OneDrive\Projects\go-tweets\docs\print
.\export-apostila-pdf.ps1
```

O script:

- usa `GO-TWEETS-APOSTILA-PROFISSIONAL-A4.md` como entrada padrao
- gera um HTML intermediario
- tenta gerar o PDF automaticamente com `wkhtmltopdf` ou `WeasyPrint`, se estiverem instalados
- aceita parametros para exportar outros materiais, como o sumario executivo

Se voce ja tinha gerado o PDF antes da modernizacao do projeto, recomendo exportar novamente para incluir:

- contrato de erro padronizado
- bootstrap atualizado
- testes automatizados

Se houver necessidade, voce tambem pode informar nomes de saida:

```powershell
.\export-apostila-pdf.ps1 -OutputPdf "apostila-go-tweets.pdf"
```

Exemplo para exportar o sumario executivo:

```powershell
.\export-apostila-pdf.ps1 `
  -InputMarkdown "GO-TWEETS-SUMARIO-EXECUTIVO-A4.md" `
  -OutputHtml "GO-TWEETS-SUMARIO-EXECUTIVO-A4.html" `
  -OutputPdf "GO-TWEETS-SUMARIO-EXECUTIVO-A4.pdf" `
  -DocumentTitle "Go Tweets Sumario Executivo" `
  -FooterTitle "Go Tweets | Sumario Executivo"
```

## Observacao sobre numeracao de pagina

A folha A4 e as quebras de pagina estao configuradas no CSS embutido.
A numeracao de pagina agora foi configurada em dois niveis:

- no CSS de impressao, para renderizadores que suportam `@page`
- no `wkhtmltopdf`, por meio de footer automatico com `pagina atual / total`

Ao exportar com o script `export-apostila-pdf.ps1` e `wkhtmltopdf`, o PDF passa a sair com rodape numerado de forma mais confiavel.

Se o renderizador usado nao suportar todas as regras visuais, ainda assim o arquivo continuara pronto para:

- tamanho A4
- margens de impressao
- quebra por capitulo
- exportacao organizada em PDF
- numeracao de pagina no PDF gerado com `wkhtmltopdf`
