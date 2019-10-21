package main
import (
    "fmt"
    "time"

    "github.com/prometheus/common/model"
    "github.com/prometheus/prometheus/pkg/labels"
    "github.com/prometheus/prometheus/promql"
)

func process_labelmatchers(LabelMatchers []*labels.Matcher) {
    fmt.Printf("Len of LabelMatchers %d\n", len(LabelMatchers))

    for i := 0; i < len(LabelMatchers); i++ {
        fmt.Printf("(%d): Type(%d) Name(%s) Value(%s)\n", i, LabelMatchers[i].Type, LabelMatchers[i].Name, LabelMatchers[i].Value)
    }
}

func process(node promql.Node, idx *uint8) {
    //fmt.Println(node.String())

    switch n:= node.(type) {
        case *promql.EvalStmt:
            fmt.Println("EvalStmt")
            process(n.Expr, idx)
        case *promql.Call:
            fmt.Printf("---%d Call %s\n", *idx, n.Func.Name)
	    *idx = *idx + 1
            process(n.Args, idx)
        case promql.Expressions:
            for _, e := range n {
               process(e, idx)
            }
        case *promql.AggregateExpr:
            fmt.Printf("---%d AggregateExpr %s\n", *idx, fmt.Sprintf("%s",n.Op))
    	    *idx = *idx + 1
            process(n.Expr, idx)
        case *promql.BinaryExpr:
            process(n.LHS, idx)

            fmt.Printf("---%d BinaryExpr %s\n", *idx, fmt.Sprintf("%s",n.Op))
	     *idx = *idx + 1

            process(n.RHS, idx)
        case *promql.VectorSelector:
            fmt.Printf("---%d VectorSelector Name %s\n", *idx, n.Name)
/*
            m := &labels.Matcher{
                Type: 0,
                Name: "name",
                Value: "val",
            }

            n.LabelMatchers = append(n.LabelMatchers, m)
            process_labelmatchers(n.LabelMatchers)
*/

    	    *idx = *idx + 1
	    if (len(n.LabelMatchers) > 1) {
                fmt.Printf("---%d VectorSelector labelMatchers %s\n", *idx, n.LabelMatchers[0].Name)
	        *idx = *idx + 1
	    } 
        case *promql.MatrixSelector:
            fmt.Printf("---%d MatrixSelector Name %s\n", *idx, n.Name)
/*
            m := &labels.Matcher{
                Type: 0,
                Name: "name",
                Value: "val",
            }

            n.LabelMatchers = append(n.LabelMatchers, m)
            process_labelmatchers(n.LabelMatchers)
*/

	    *idx = *idx + 1
	    if (len(n.LabelMatchers) > 1) {
                fmt.Printf("---%d MatrixSelector labelMatchers %s\n", *idx, n.LabelMatchers[0].Name)
	        *idx = *idx + 1
	    }
	    if (n.Range != time.Duration(0)) {
                fmt.Printf("---%d MatrixSelector Range %s\n", *idx, fmt.Sprint(model.Duration(n.Range)))
	        *idx = *idx + 1
	    }
        case *promql.UnaryExpr:
            fmt.Printf("---%d UnaryExpr\n", *idx)
            process(n.Expr, idx)
        case *promql.ParenExpr:
            fmt.Println("ParenExpr")
            process(n.Expr, idx)
        case *promql.NumberLiteral:
            fmt.Printf("---%d NumberLiteral %s\n", *idx, fmt.Sprint(n.Val))

	    if *idx == 12 {
		n.Val = 75
	    }
	    if *idx == 25 {
		n.Val = 85
	    }

	    *idx = *idx + 1
    }
}

func main() {

//    expr, err := promql.ParseExpr("sum(go_memstats_alloc_bytes{namespace='nm1'})/sum(go_memstats_stack_inuse_bytes)*100")
//      expr, err := promql.ParseExpr("rate(ldap_sum[5m])/rate(ldap_count[5m])")
      expr, err := promql.ParseExpr("((node_filesystem_size_bytes{device=\"rootfs\"} - node_filesystem_free_bytes{device=\"rootfs\"}) / node_filesystem_size_bytes{device=\"rootfs\"} * 100 > 70) and ((node_filesystem_size_bytes{device=\"rootfs\"} - node_filesystem_free_bytes{device=\"rootfs\"}) / node_filesystem_size_bytes{device=\"rootfs\"} * 100 <= 80)")

    if err != nil {
        fmt.Println("err : %s", err)
        return
    }

    fmt.Println(expr.String())

    fmt.Println(promql.Tree(expr))

    idx := uint8(1)
    process(expr, &idx)

    fmt.Println(expr.String())
}
