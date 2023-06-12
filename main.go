package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/correiostech/rastro"
	"github.com/correiostech/token"
)

var (
	urlToken  = flag.String("tk", "https://api.correios.com.br/token/v1/autentica", "url para obter token")
	urlAsync  = flag.String("e", "https://api.correios.com.br/rastro-async/v1/objetos/async?resultado=U", "endpoint de rastro")
	urlRecibo = flag.String("r", "https://api.correios.com.br/rastro-async/v1/recibo/", "url para checar recibo")
	usuario   = flag.String("u", "", "nome de usuario de acesso às API dos Correios")
	senha     = flag.String("p", "", "senha do usuario de acesso às API dos Correios")
	file      = flag.String("a", "objetos.txt", "nome do arquivo a ser lido")
	tempo     = flag.Int("s", 2, "tempo para execucao, padrão 15 segundos")
)

func formataData(data string) (string, error) {
	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, data)
	if err != nil {
		return "", err
	}
	brLayout := "02/01/2006 15:04:05"
	brStr := t.Format(brLayout)
	return brStr, nil
}

func leArquivo(arquivo string) ([][]string, error) {
	//slice que irá receber cada um objeto
	objetos := make([][]string, 0)

	//lê o arquivo
	dados, err := ioutil.ReadFile(arquivo)
	if err != nil {
		return nil, err
	}
	conteudo := string(dados)

	//divide os objetos em grupos de 1000
	linhas := strings.Split(conteudo, "\n")
	for i := 0; i < len(linhas); i += 1000 {
		fim := i + 1000
		if fim > len(linhas) {
			fim = len(linhas)
		}
		objetos = append(objetos, linhas[i:fim])
	}
	return objetos, nil
}

func init() {
	*usuario = os.Getenv("USUARIO_API_CORREIOS")
	*senha = os.Getenv("SENHA_API_CORREIOS")
}

func rastreiaAsync(objetos []string) (string, error) {
	clientRastro, err := rastro.New(*urlAsync, token.Token)
	if err != nil {
		return "", fmt.Errorf("erro: rastreia objetos: " + err.Error())
	}
	res, err := clientRastro.RastreiaAsync(objetos, token.Token)
	if err != nil {
		return "", err
	}
	return res.Numero, nil
}

func checaRecibo(recibo string) (rastro.Resultado, error) {
	var res rastro.Resultado
	clientRecibo, err := rastro.New(*urlRecibo, token.Token)
	if err != nil {
		return res, fmt.Errorf("erro: checa recibo 1: " + err.Error())
	}
	res, err = clientRecibo.Recibo(recibo, token.Token)
	if err != nil {
		return res, fmt.Errorf("erro: checa recibo 2: " + err.Error())
	}
	return res, nil
}

func main() {
	flag.Parse()
	clientToken, err := token.GetToken(*urlToken, *usuario, *senha)
	if err != nil {
		log.Panic(err.Error())
	}
	token.Token = clientToken.Token
	objetos, err := leArquivo(*file)
	if err != nil {
		log.Panic(err.Error())
	}
	var recibos []string
	for _, obj := range objetos {
		recibo, err := rastreiaAsync(obj)
		if err != nil {
			log.Panic(err.Error())
		}
		log.Println("objetos registrados, recibo:", recibo)
		recibos = append(recibos, recibo)
	}
	for _, rec := range recibos {
		rastros, err := checaRecibo(rec)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		for _, o := range rastros.Objetos {
			dt, err := formataData(o.Eventos[0].DataHora)
			if err != nil {
				log.Println(err.Error())
			}
			fmt.Printf("%s: %s %s %s %s\n",
				dt,
				o.CodigoObjeto,
				o.Eventos[0].Codigo,
				o.Eventos[0].Tipo,
				o.Eventos[0].Descricao,
			)
		}
	}
}
